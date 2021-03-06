package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

	"template/internal/service"
	"template/internal/store"
	"template/pkg/infra/mongo"
	"template/pkg/infra/mysql"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/web"
	libKVStore "github.com/docker/libkv/store"
	"github.com/go-redis/redis/v8"
	"github.com/google/gops/agent"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	serverName = "svr"

	consulAddrKey = "consul"
	consulAddrDef = "127.0.0.1:8500"

	logLevelKey = "loglevel"
	logLevelDef = "info"

	printVersionKey = "version"
	printVersionDef = false

	consulPrefixKey = "prefix"
	consulPrefixDef = ""
)

func init() {
	// NOTE: go-micro 只支持小写字母的选项
	flag.String(consulAddrKey, consulAddrDef, "the consul address")
	flag.String(logLevelKey, logLevelDef, "log level")
	flag.String(consulPrefixKey, consulPrefixDef, "consul key prefix")
	flag.Bool(printVersionKey, printVersionDef, "print program build version")

	flag.Parse()
}

// App ...
type App interface {
	Run(chan<- os.Signal) error
	Stop() error
}

// New ...
func New(options ...Option) (App, error) {
	svc := &app{ctx: context.Background()}
	for _, opt := range options {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

type app struct {
	nodeID     int
	rpcService micro.Service
	webService web.Service
	useCase    service.UseCase
	conf       config.Config
	redisCli   redis.UniversalClient
	mysqlCli   mysql.Client
	mongoCli   mongo.Client
	dao        store.Dao
	kvStore    libKVStore.Store
	ctx        context.Context
}

// Run ...
func (a *app) Run(ch chan<- os.Signal) error {

	if err := agent.Listen(agent.Options{
		ConfigDir:       os.TempDir(),
		ShutdownCleanup: true,
	}); err != nil {
		return err
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		if a.rpcService == nil {
			return nil
		}

		if err := a.rpcService.Run(); err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		if a.webService == nil {
			return nil
		}

		if err := a.webService.Run(); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		ch <- syscall.SIGQUIT
		return err
	}
	return nil
}

// Stop ...
func (a *app) Stop() error {
	return nil
}

func (a *app) intranetIP() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addr {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil && ipNet.IP.IsGlobalUnicast() {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("valid local IP not found")
}

func (a *app) makeConsulKey(key string) string {
	keyPrefix := a.conf.Get(consulPrefixKey).String(consulPrefixDef)
	if keyPrefix == "" {
		return fmt.Sprintf("%v/%v", serverName, key)
	}
	return fmt.Sprintf("%v/%v/%v", keyPrefix, serverName, key)
}

func (a *app) getConsulConf(key string, data interface{}, def interface{}) error {
	consulKey := a.makeConsulKey(key)
	kvPair, err := a.kvStore.Get(consulKey)
	if err != nil {
		if err != libKVStore.ErrKeyNotFound {
			return err
		}

		// first startup
		value, err := json.MarshalIndent(def, "", "\t")
		if err != nil {
			return err
		}

		_, kvPair, err = a.kvStore.AtomicPut(consulKey, value, nil, &libKVStore.WriteOptions{IsDir: false})
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(kvPair.Value, data)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) watchConsulConf(key string, observer ConfigObserver) error {
	kvChan, err := a.kvStore.Watch(a.makeConsulKey(key), make(chan struct{}, 1))
	if err != nil {
		return err
	}

	go func() {
		for kv := range kvChan {
			if err = observer.OnConfigChanged(key, kv.Value); err != nil {
				log.Err(err).Str("key", key).Str("value", string(kv.Value)).Msg("watch consul key")
			}
		}
	}()

	return nil
}

func (a *app) watchConsulConfTree(root string, observer ConfigObserver) error {
	kvChan, err := a.kvStore.WatchTree(a.makeConsulKey(root), make(chan struct{}, 1))
	if err != nil {
		return err
	}

	go func() {
		for kv := range kvChan {
			for _, pair := range kv {
				if pair.Value == nil {
					continue
				}
				idx := strings.LastIndex(pair.Key, root) + len(root)
				if err = observer.OnConfigChanged(strings.TrimPrefix(pair.Key[idx:], "/"), pair.Value); err != nil {
					log.Err(err).Str("key", pair.Key).Str("value", string(pair.Value)).Msg("watch consul dir")
				}
			}
		}
	}()

	return nil
}
