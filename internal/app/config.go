package app

import (
	"encoding/json"
	"fmt"
	"sync"
)

const (
	redisModeCluster = "cluster"
	// redisModeStandalone = "standalone"

	bizConfKey = "biz"
)

type ConfigObserver interface {
	OnConfigChanged(key string, data []byte)
}

type ConfigHandler func(key string, data []byte)

func (f ConfigHandler) OnConfigChanged(key string, data []byte) {
	f(key, data)
}

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

type bizConf struct {
	rw              sync.RWMutex
	ActiveBeginTime string `json:"activeBeginTime"`
}

func (biz *bizConf) OnConfigChanged(key string, data []byte) {
	biz.rw.Lock()
	defer biz.rw.Unlock()

	switch key {
	case bizConfKey:
		_ = json.Unmarshal(data, biz)
	}
}

func (biz *bizConf) GetActiveBeginTime() string {
	biz.rw.RLock()
	defer biz.rw.RUnlock()

	return biz.ActiveBeginTime
}
