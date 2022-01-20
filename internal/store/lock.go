package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
)

const (
	dLockPrefix = "ffa:game:lock:{%v}"

	// redis-cli --eval delLockLua.lua lock:user , haha
	delLockLua = `
		local key = KEYS[1]
		local token = ARGV[1]
		
		local value = redis.call("GET", key)
		if value ~= token then
			return 0
		end

		return redis.call("DEL", key)
	`

	// redis-cli -c -p 7000  --eval multiLock.lua {lock}:ddd {lock}:ccc , haha 100000
	multiLockLua = `
		local token = ARGV[1]
		local expire = ARGV[2]

		for i, v in ipairs(KEYS) do
			if (redis.call("SET", v, token, "NX", "PX", expire) == false) then
				for k=i-1,1,-1 do
					redis.call("DEL", KEYS[k])
				end
				return false
			end
		end

		return "OK"
	`

	// redis-cli -c -p 7000 --eval delMultiLock.lua {lock}:ddd {lock}:ccc , haha
	delMultiLockLua = `
		local token = ARGV[1]
		local count = 0

		for _, v in ipairs(KEYS) do
			if (redis.call("get", v) ~= token) then
				break
			end
			count = count + redis.call("del",v) 
		end

		return count
	`
)

type Lock interface {
	NewDistLock(key string) DistLock
}

// DistLock 分布式锁
type DistLock interface {
	UnLock() bool
	Lock() bool
	TryLock(tryTimes, milliSleep int) bool
}

// NewDistLock ...
func (d *daoImpl) NewDistLock(key string) DistLock {
	return &redisDistLock{
		key:   fmt.Sprintf(dLockPrefix, key),
		token: xid.New().String(),
		cli:   d.redisRepo,
	}
}

type redisDistLock struct {
	key   string
	token string
	cli   redis.UniversalClient
}

// UnLock ...
func (r *redisDistLock) UnLock() bool {
	value, err := r.cli.Eval(context.TODO(), delLockLua, []string{r.key}, r.token).Int()
	if err != nil {
		return false
	}

	return value != 0
}

// Lock ...
func (r *redisDistLock) Lock() bool {
	return r.getRedisLock(r.key, r.token)
}

// TryLock ...
func (r *redisDistLock) TryLock(tryTime, milliSleep int) (ok bool) {
	for i := 0; i < tryTime; i++ {
		ok = r.getRedisLock(r.key, r.token)
		if ok {
			return
		}

		time.Sleep(time.Duration(milliSleep) * time.Millisecond)
	}

	return
}

func (r *redisDistLock) getRedisLock(key string, token string) bool {
	const dLockExpire = 1000 // ms
	value, err := r.cli.SetNX(context.TODO(), key, token, time.Millisecond*dLockExpire).Result()
	if err != nil {
		return false
	}
	return value
}
