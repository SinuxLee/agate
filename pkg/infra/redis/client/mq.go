package client

import (
	"errors"
	"sync"

	"github.com/gomodule/redigo/redis"
)

const (
	// UserStatusChannel 通知玩家登录登出状态
	UserStatusChannel = "{ch}/user/status" // center -> room
)

// Topic 订阅主题
type Topic struct {
	ChannelName      string
	IsChannelPattern bool                                                            //频道名支持正则
	MsgHandler       func(channelName string, channelPattern string, message []byte) //获取到了channel的消息
	EventHandler     func(channelName string, eventName string, consumerCount int)   //订阅、取消订阅等事件
	ErrorHandler     func(topic *Topic, err error)                                   //出错回调
	pubSubCon        *redis.PubSubConn
	stopFlag         bool //为true是停止
}

// RedisMQ ...
type RedisMQ interface {
	Publish(topic string, msg []byte) (err error)
	Subscribe(topic Topic) (err error)
	Unsubscribe(topicName string) bool
}

// NewRedisMQ ...
func NewRedisMQ(cli RedisClient) RedisMQ {
	if cli == nil {
		return nil
	}

	entity := &redisMQ{client: cli}
	entity.topics = make(map[string]*Topic)

	return entity
}

type redisMQ struct {
	sync.RWMutex
	client RedisClient
	topics map[string]*Topic
}

func (r *redisMQ) Publish(topic string, msg []byte) (err error) {
	con := r.client.getConnection()
	if con.Err() != nil {
		err = con.Err()
		return
	}
	defer con.Close()

	if _, err = redis.Int(con.Do("PUBLISH", topic, msg)); err != nil {
		return
	}

	return
}

func (r *redisMQ) Subscribe(topic Topic) (err error) {
	if len(topic.ChannelName) == 0 {
		err = errors.New("channel name is empty")
		return
	}

	con := r.client.getConnection()
	if con.Err() != nil {
		err = con.Err()
		return
	}

	psCon := redis.PubSubConn{Conn: con}
	if topic.IsChannelPattern {
		err = psCon.PSubscribe(topic.ChannelName)
	} else {
		err = psCon.Subscribe(topic.ChannelName)
	}
	if err != nil {
		psCon.Close()
		return
	}

	topic.pubSubCon = &psCon
	r.Lock()
	r.topics[topic.ChannelName] = &topic
	r.Unlock()

	go r.defaultMessageHandler(&topic)

	return
}

func (r *redisMQ) Unsubscribe(topicName string) bool {
	r.Lock()
	defer r.Unlock()
	topic, ok := r.topics[topicName]
	if !ok {
		return false
	}

	topic.stopFlag = true

	return true
}

func (r *redisMQ) defaultMessageHandler(topic *Topic) {
	for !topic.stopFlag {
		switch v := topic.pubSubCon.ReceiveWithTimeout(0).(type) {
		case redis.Message:
			if topic.MsgHandler != nil {
				topic.MsgHandler(v.Channel, v.Pattern, v.Data)
			}
		case redis.Subscription:
			if topic.EventHandler != nil {
				topic.EventHandler(v.Channel, v.Kind, v.Count)
			}
		case error: //出错
			if topic.ErrorHandler != nil {
				topic.ErrorHandler(topic, v)
			}
		}
	}

	if topic.IsChannelPattern {
		_ = topic.pubSubCon.PUnsubscribe(topic.ChannelName)
	} else {
		_ = topic.pubSubCon.Unsubscribe(topic.ChannelName)
	}

	topic.pubSubCon.Close()

	r.Lock()
	delete(r.topics, topic.ChannelName)
	r.Unlock()
}
