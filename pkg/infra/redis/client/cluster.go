package client

import (
	infraRedis "template/pkg/infra/redis"

	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

func NewCluster(conf Config) RedisClient {
	cluster, err := infraRedis.NewCluster(conf.Address, conf.Password, &conf.PoolConfig)
	if err != nil {
		return nil
	}
	return &clusterClient{
		cluster: cluster,
	}
}

type clusterClient struct {
	cluster *redisc.Cluster
}

func (c *clusterClient) ExecuteCommand(cmd string, args ...interface{}) (interface{}, error) {
	// 获取连接实例
	con := c.cluster.Get()
	if con.Err() != nil {
		return nil, con.Err()
	}
	defer con.Close()
	res, errDo := con.Do(cmd, args...)
	return res, errDo
}

func (c *clusterClient) ExecuteScript(cmd string, args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (c *clusterClient) getConnection() redis.Conn {
	return c.cluster.Get()
}
