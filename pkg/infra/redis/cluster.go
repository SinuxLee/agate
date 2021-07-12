package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

func NewCluster(addr string, password string, cfg *PoolConfig) (*redisc.Cluster, error) {
	cluster := &redisc.Cluster{
		StartupNodes: []string{addr},
		DialOptions: []redis.DialOption{redis.DialPassword(password),
			redis.DialConnectTimeout(3 * time.Second),
			redis.DialReadTimeout(3 * time.Second),
			redis.DialWriteTimeout(3 * time.Second)},
		CreatePool: func(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
			return &redis.Pool{
				MaxIdle:         cfg.MaxIdle,
				MaxActive:       cfg.MaxActive,
				Wait:            cfg.Wait,
				IdleTimeout:     cfg.IdleTimeout,
				MaxConnLifetime: cfg.MaxConnLifetime,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("tcp", addr, opts...)
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					if time.Since(t) < time.Minute {
						return nil
					}
					_, err := c.Do("PING")
					return err
				},
			}, nil
		},
	}

	err := cluster.Refresh()
	if err != nil {
		return nil, err
	}

	conn := cluster.Get()
	err = conn.Err()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return cluster, nil
}
