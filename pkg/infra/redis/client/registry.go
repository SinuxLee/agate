package client

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

const (
	// RegistryExpireTime 服务注册信息有效期
	RegistryExpireTime = 1 * time.Minute

	// ServerTypeBase ...
	ServerTypeBase = 10000
	serverTypeSep  = ":"
	registryKey    = "module:server"
)

// ServerInfo ...
type ServerInfo struct {
	IP         string
	Port       uint16
	ServerType int
	NodeID     int
}

// FromString ...
func (info *ServerInfo) FromString(s string) (err error) {
	str := strings.Split(s, serverTypeSep)
	if len(str) < 3 {
		err = errors.New("split failed")
		return
	}

	if addr := net.ParseIP(str[0]); addr == nil {
		err = errors.New("parse ip failed:" + str[0])
		return
	}

	info.IP = str[0]

	port, err := strconv.Atoi(str[1])
	if err != nil {
		err = errors.New("port parse failed:" + str[1])
		return
	}

	info.Port = uint16(port)

	if info.Port == 0 {
		err = errors.New("port can't be zero")
		return
	}

	id, err := strconv.Atoi(str[2])
	if err != nil {
		err = errors.New("convert to int failed:" + str[2])
		return
	}

	info.ServerType = id / ServerTypeBase
	info.NodeID = id % ServerTypeBase

	if info.ServerType <= 0 {
		err = errors.New("server type must be greater than 0")
		return
	}

	if info.NodeID <= 0 {
		err = errors.New("node id must be greater than 0")
		return
	}

	return
}

// String ...
func (info *ServerInfo) String() (str string, err error) {
	if addr := net.ParseIP(info.IP); addr == nil {
		err = errors.New("parse ip failed:" + info.IP)
		return
	}

	buff := bytes.NewBufferString(info.IP)
	buff.WriteString(serverTypeSep)

	if info.Port == 0 {
		err = errors.New("port can't be zero")
		return
	}

	buff.WriteString(strconv.Itoa(int(info.Port)))
	buff.WriteString(serverTypeSep)

	if info.ServerType <= 0 {
		err = errors.New("server type must be greater than 0")
		return
	}

	if info.NodeID <= 0 {
		err = errors.New("node id must be greater than 0")
		return
	}

	buff.WriteString(strconv.Itoa(info.ServerType*ServerTypeBase + info.NodeID))

	str = buff.String()
	return
}

// Registry ...
type Registry interface {
	Register(info ServerInfo) error
	Deregister(info ServerInfo) error
	GetService(t int) []ServerInfo
	ListServices() []ServerInfo
}

// NewRegistry ...
func NewRegistry(cli RedisClient) Registry {
	if cli == nil {
		return nil
	}

	reg := &registry{
		client: cli,
		infos:  make(map[string]int64),
	}

	go reg.checkExpire()

	return reg
}

type registry struct {
	sync.RWMutex
	client RedisClient
	infos  map[string]int64
}

// Register ...
func (r *registry) Register(info ServerInfo) (err error) {
	nodeID, err := redis.Int64(r.client.ExecuteCommand("incr", fmt.Sprintf("serverType:%d:nodoId", info.ServerType)))
	if err != nil {
		return
	}

	info.NodeID = int(nodeID+100) % ServerTypeBase //加100是为了不与线上ini的配置产生冲突
	str, err := info.String()
	if err != nil {
		return
	}

	r.RLock()
	_, exist := r.infos[str]
	r.RUnlock()
	if exist {
		return
	}

	timestamp := time.Now().Unix()
	_, err = redis.Bool(r.client.ExecuteCommand("hset", registryKey, str, timestamp))
	if err != nil {
		return
	}

	r.Lock()
	r.infos[str] = timestamp
	r.Unlock()

	return
}

// Deregister ...
func (r *registry) Deregister(info ServerInfo) (err error) {
	str, err := info.String()
	if err != nil {
		return
	}

	r.RLock()
	_, exist := r.infos[str]
	r.RUnlock()
	if !exist {
		return
	}

	_, err = redis.Bool(r.client.ExecuteCommand("hdel", registryKey, str))
	if err != nil {
		return
	}

	r.Lock()
	delete(r.infos, str)
	r.Unlock()

	return
}

// GetService ...
func (r *registry) GetService(t int) (l []ServerInfo) {
	if t <= 0 {
		return
	}

	list := r.ListServices()
	if len(list) == 0 {
		return
	}

	for _, value := range list {
		if value.ServerType == t {
			if l == nil {
				l = make([]ServerInfo, 0)
			}

			l = append(l, value)
		}
	}

	return
}

// ListServices ...
func (r *registry) ListServices() (l []ServerInfo) {
	list, err := redis.StringMap(r.client.ExecuteCommand("hgetall", registryKey))
	if err == nil {
		idx := 0
		for key := range list {
			info := new(ServerInfo)
			err := info.FromString(key)
			if err == nil {
				if l == nil {
					l = make([]ServerInfo, 0)
				}

				l = append(l, *info)
				idx++
			}
		}
	}

	return
}

func (r *registry) checkExpire() {
	timer := time.NewTimer(time.Second)
	var nextKey string
	for range timer.C {
		// Note: case里的defer不生效
		r.Lock()
		if len(r.infos) == 0 {
			timer.Reset(time.Second)
			r.Unlock()
			break
		}

		if _, exist := r.infos[nextKey]; exist {
			if err := r.updateExpireTime(nextKey); err != nil {
				log.Error().Str("key", nextKey).Err(err).Msg("update expire time failed")
			} else {
				log.Info().Str("key", nextKey).Msg("update key's expire time in redis")
			}
		}

		var smaller int64 = math.MaxInt64
		for key, value := range r.infos {
			if value < smaller {
				nextKey = key
				smaller = value
			}
		}

		expire := RegistryExpireTime - time.Duration(time.Now().Unix()-smaller)*time.Second
		timer.Reset(expire)
		r.Unlock()
	}
}

func (r *registry) updateExpireTime(key string) (err error) {
	timestamp := time.Now().Unix()
	_, err = redis.Bool(r.client.ExecuteCommand("hset", registryKey, key, timestamp))
	if err != nil {
		return
	}

	r.infos[key] = timestamp

	return
}
