package redis

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// PoolConfig ...
type PoolConfig struct {
	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration
}

// NewScript ...
func NewScript(keyCount int, src string) *Script {
	return &Script{script: redis.NewScript(keyCount, src)}
}

// Pool redis.Pool封装，包含trace等实现
type Pool struct {
	pool *redis.Pool
}

// Script redis.Script封装，包含trace等实现
type Script struct {
	script *redis.Script
}

// Do 兼容redigo
func (s *Script) Do(c redis.Conn, keysAndArgs ...interface{}) (interface{}, error) {
	return s.script.Do(c, keysAndArgs...)
}

func parseRedisErr(err error) string {
	if err == nil {
		return "OK"
	}
	if err == redis.ErrNil {
		return "OK"
	}
	switch e := err.(type) {
	case *net.OpError:
		// net error
		return e.Op + ": " + e.Err.Error()

	}
	return err.Error()
}

// Get 兼容redigo
func (p *Pool) Get() redis.Conn {
	return p.pool.Get()
}

func (p *Pool) Do(cmd string, args ...interface{}) (interface{}, error) {
	// 获取连接实例
	con := p.pool.Get()
	defer con.Close()
	res, errDo := con.Do(cmd, args...)
	return res, errDo
}

// NewPool1 创建Redis连接池管理对象
func NewPool(addr string, passwd string, cfg *PoolConfig) *Pool {
	pool := &redis.Pool{
		MaxIdle:         cfg.MaxIdle,
		MaxActive:       cfg.MaxActive,
		Wait:            cfg.Wait,
		IdleTimeout:     cfg.IdleTimeout,
		MaxConnLifetime: cfg.MaxConnLifetime,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr,
				redis.DialPassword(passwd),
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	if nil != conn.Err() {
		return nil
	}
	defer conn.Close()

	return &Pool{pool: pool}
}

func joinArgs(args ...interface{}) string {
	var paramSlice []string
	for _, param := range args {
		paramSlice = append(paramSlice, fmt.Sprintf("%v", param))
	}
	aa := strings.Join(paramSlice, ",") // Join 方法第2个参数是 string 而不是 rune
	return aa
}
