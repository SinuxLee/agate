package client

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/xid"
)

const (
	lockPrefix = "lock:"
	lockExpire = 1000 //ms

	delLockLua = `
		local key = KEYS[1]
		local argv = ARGV[1]
		
		local value = redis.call("GET", key)
		
		if value == argv then
			return redis.call("DEL", key)
		else
			return 0
		end
	`
)

var lockScript *redis.Script

func init() {
	lockScript = redis.NewScript(1, delLockLua)
}

// RedisDistLock redis 分布式锁
type RedisDistLock interface {
	UnLock() error
	Lock() bool
	TryLock(tryTime, milliSleep int) bool
}

// NewDistLock ...
func NewDistLock(c RedisClient, key string) RedisDistLock {
	return &redisDistLock{
		key:   lockPrefix + key,
		token: xid.New().String(),
		cli:   c,
	}
}

type redisDistLock struct {
	key   string
	token string
	cli   RedisClient
}

// UnLock ...
func (r *redisDistLock) UnLock() error {
	return releaseRedisLock(r.cli, r.key, r.token)
}

// Lock ...
func (r *redisDistLock) Lock() bool {
	return getRedisLock(r.cli, r.key, r.token)
}

// TryLock ...
func (r *redisDistLock) TryLock(tryTime, milliSleep int) bool {
	ret := getRedisLock(r.cli, r.key, r.token)
	for i := 0; i < tryTime; i++ {
		if ret {
			return ret
		}
		time.Sleep(time.Duration(milliSleep) * time.Millisecond)
		ret = getRedisLock(r.cli, r.key, r.token)
	}
	return ret
}

func getRedisLock(cli RedisClient, key string, token string) bool {
	value, err := redis.String(cli.ExecuteCommand("SET", key, token, "NX", "PX", lockExpire))
	if err != nil {
		return false
	}
	return value == "OK"
}

func releaseRedisLock(cli RedisClient, key string, token string) error {
	con := cli.getConnection()
	err := con.Err()
	if err == nil {
		return err
	}
	defer con.Close()

	_, err = lockScript.Do(con, key, token)

	return err
}
