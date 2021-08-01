package app

import "fmt"

const (
	redisModeCluster = "cluster"
	// redisModeStandalone = "standalone"
)

type redisConf struct {
	Mode     string `json:"mode"`
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

func (r *redisConf) Info() string {
	return fmt.Sprintf("mode:%v addr:%v password:%v", r.Mode, r.Addr, r.Password)
}

type mysqlConf struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func (m *mysqlConf) Info() string {
	return fmt.Sprintf("host:%v port:%v user:%v password:%v db:%v",
		m.Host, m.Port, m.User, m.Password, m.Database)
}

type mongodbConf struct {
	Host     []string `json:"host"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	Database string   `json:"database"`
}

func (m *mongodbConf) Info() string {
	return fmt.Sprintf("host:%v user:%v password:%v db:%v",
		m.Host, m.User, m.Password, m.Database)
}

type webConf struct {
	GinMode string `json:"ginMode"`
	Port    string `json:"port"`
}

type rpcConf struct {
	Port string `json:"port"`
}
