package client

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

const (
	// UserStateQueue 玩家房间状态队列
	UserStateQueue = "{ch}/CenterServer/Listen" // room -> center

	// AssetChangeQueue 资产变更队列
	AssetChangeQueue = "{ch}/user/asset/change" // room -> center

	// AssetChangeResQueue 资产变更结果
	AssetChangeResQueue = "{ch}/user/asset/change_res" // center -> room
)

/**
example: push string "Hello" to queue
				————————————————————
push left ->	|o|l|l|e|H| | | | |		-> pop right
				____________________
*/

// RedisQueue ...
type RedisQueue interface {
	PushData(queueName string, data []byte) (size int, err error)
	PopDataBlock(queueName string) (data []byte, err error)
	PopDataBlockWithTimeout(queueName string, timeout int) (data []byte, err error) //timeout不能超过3
	PopDataNoBlock(queueName string) (data []byte, err error)
}

// NewRedisQueue timeout为等待时间(单位:秒),为0时一直等待.
func NewRedisQueue(cli RedisClient, timeout int) RedisQueue {
	if cli == nil {
		return nil
	}
	redq := &redisQueue{client: cli, timeout: timeout}

	return redq
}

type redisQueue struct {
	client  RedisClient
	timeout int
}

func (r *redisQueue) PushData(queueName string, data []byte) (size int, err error) {
	if len(queueName) == 0 {
		err = errors.New("queue name is empty")
		return
	}

	if len(data) == 0 {
		err = errors.New("data is empty")
		return
	}

	con := r.client.getConnection()
	if con.Err() != nil {
		err = con.Err()
		return
	}

	if result, rdErr := redis.Int(con.Do("LPUSH", queueName, data)); rdErr != nil {
		err = rdErr
	} else {
		size = result
	}

	_ = con.Close()
	return
}

func (r *redisQueue) PopDataBlock(queueName string) (data []byte, err error) {
	return r.PopDataBlockWithTimeout(queueName, r.timeout)
}

func (r *redisQueue) PopDataBlockWithTimeout(queueName string, timeout int) (data []byte, err error) {
	if len(queueName) == 0 {
		err = errors.New("queue name is empty")
		return
	}
	if timeout <= 0 {
		timeout = r.timeout
	}
	con := r.client.getConnection()
	if con.Err() != nil {
		err = con.Err()
		return
	}

	if result, rdErr := redis.Values(con.Do("BRPOP", queueName, timeout)); rdErr != nil {
		err = rdErr
	} else {
		if len(result) >= 2 {
			if name, ok := result[0].([]byte); ok && string(name) == queueName {
				if data, ok = result[1].([]byte); !ok {
					err = errors.New("can't convert data to []byte")
				}
			}
		}
	}

	_ = con.Close()

	return
}

func (r *redisQueue) PopDataNoBlock(queueName string) (data []byte, err error) {
	if len(queueName) == 0 {
		err = errors.New("queue name is empty")
		return
	}

	con := r.client.getConnection()
	if con.Err() != nil {
		err = con.Err()
		return
	}

	if result, rdErr := redis.Bytes(con.Do("RPOP", queueName)); rdErr != nil {
		err = rdErr
	} else {
		data = result
	}

	_ = con.Close()
	return
}
