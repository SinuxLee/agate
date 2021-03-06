package monitoring

import (
	"errors"
	"net/http"

	"github.com/imdario/mergo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Addr       string `json:"addr"`
	Path       string `json:"path"`
	ServerName string `json:"-"`
}

var (
	defaultConf = &Config{
		Addr:       "",
		Path:       "",
		ServerName: "",
	}
)

func Serve(conf *Config) error {
	if err := mergo.Merge(defaultConf, conf); err != nil {
		return err
	}

	if defaultConf.Addr == "" || defaultConf.Path == "" {
		return errors.New("invalid prometheus config")
	}

	initMysql()
	initRedis()
	// 处理监听问题
	http.Handle(defaultConf.Path, promhttp.Handler())

	go func() {
		_ = http.ListenAndServe(defaultConf.Addr, nil)
	}()

	return nil
}
