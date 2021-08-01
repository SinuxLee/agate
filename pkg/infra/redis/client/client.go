// Package client ...
package client

import (
	infraRedis "template/pkg/infra/redis"

	"github.com/gomodule/redigo/redis"
)

const (
	// RedisDoSuccess ...
	RedisDoSuccess = "OK"
)

// Config ...
type Config struct {
	infraRedis.PoolConfig
	Address  string
	Password string
	DbIndex  int
}

// RedisClient ...
type RedisClient interface {
	ExecuteCommand(string, ...interface{}) (interface{}, error)
	ExecuteScript(string, ...interface{}) (interface{}, error)
	getConnection() redis.Conn
}

type redisClient struct {
	redisPoll *infraRedis.Pool
	config    Config
}

// NewClient create new redis client with Config
func NewClient(conf Config) RedisClient {
	entity := &redisClient{config: conf}
	entity.redisPoll = infraRedis.NewPool(conf.Address, conf.Password, &entity.config.PoolConfig)
	if entity.redisPoll == nil {
		return nil
	}

	c := entity.getConnection() //如果redis地址填错了，这里会阻塞
	defer func() {
		_ = c.Close()
	}()
	if replay, err := redis.String(c.Do("ping")); err != nil && replay != "PONG" {
		return nil
	}

	if conf.DbIndex != 0 {
		_, _ = entity.ExecuteCommand("select ", conf.DbIndex)
	}

	return entity
}

func (r *redisClient) getConnection() redis.Conn {
	return r.redisPoll.Get()
}

func (r *redisClient) ExecuteCommand(cmd string, args ...interface{}) (interface{}, error) {
	return r.redisPoll.Do(cmd, args...)
}

func (r *redisClient) ExecuteScript(file string, args ...interface{}) (interface{}, error) {
	return execute(file, args...)
}
