package app

import (
	"context"
	"flag"
	"os"
	"syscall"
	"template/internal/store"

	"template/internal/api/rest"
	"template/internal/service"
	"template/pkg/infra/mongo"
	"template/pkg/infra/mysql"
	"template/pkg/infra/redis/client"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/web"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func init() {
	flag.String("consul_addr", "127.0.0.1:8500", "the consul address")
	flag.String("loglevel", "info", "log level")
	flag.Parse()
}

// App ...
type App interface {
	Run(chan<- os.Signal) error
	Stop() error
}

// New ...
func New(options ...Option) (App, error) {
	// init app component
	svc := &app{}
	for _, opt := range options {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

type app struct {
	rpcService  micro.Service
	webService  web.Service
	ginRouter   *gin.Engine
	restHandler rest.Handler
	useCase     service.UseCase
	conf        config.Config
	redisCli    client.RedisClient
	mysqlCli    mysql.Client
	mongoCli    mongo.Client
	dao         store.Dao
}

// Run ...
func (a *app) Run(ch chan<- os.Signal) error {
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
