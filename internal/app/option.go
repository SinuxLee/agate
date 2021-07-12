package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"template/internal/api/rest"
	"template/internal/api/rpc"
	"template/internal/service"
	"template/internal/store"
	"template/pkg/infra/mongo"
	"template/pkg/infra/mysql"
	"template/pkg/infra/redis"
	redisClient "template/pkg/infra/redis/client"
	"template/pkg/middleware"
	"template/pkg/proto"

	"github.com/asim/go-micro/plugins/logger/zerolog/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	selectorReg "github.com/asim/go-micro/plugins/selector/registry"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	microLimiter "github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/config/source/env"
	"github.com/asim/go-micro/v3/config/source/flag"
	"github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/micro/cli/v2"
)

// Option ...
type Option func(*app) error

func Config() Option {
	return func(a *app) (err error) {
		flagSource := flag.NewSource(
			flag.IncludeUnset(true),
		)

		a.conf, err = config.NewConfig()
		if err != nil {
			return err
		}

		err = a.conf.Load(flagSource, env.NewSource())
		if err != nil {
			return err
		}

		logger.Info("Load config successfully.")
		return nil
	}
}

func Logger() Option {
	return func(a *app) error {
		level, err := logger.GetLevel(a.conf.Get("loglevel").String("info"))
		if err != nil {
			level = logger.DebugLevel
		}

		logger.DefaultLogger = zerolog.NewLogger(logger.WithOutput(os.Stdout), logger.WithLevel(level))

		logger.Info("Init logger successfully.")
		return nil
	}
}

func RpcService() Option {
	return func(a *app) error {
		a.rpcService = micro.NewService(
			micro.Server(server.NewServer(
				server.Name("greeterRpc"), // consul 中的 service name
				server.Id("1001"),
				server.Transport(grpc.NewTransport()),
				server.Metadata(map[string]string{
					"nodeId":      "1001",
					"serviceName": "greeter",
					"type":        "rpcService",
				}),
				server.Version(server.DefaultVersion),
				server.Address(":18086"),
				server.RegisterTTL(time.Second*3),
				server.RegisterInterval(time.Second),
				server.Registry(consul.NewRegistry(
					registry.Addrs("127.0.0.1:8500"),
					registry.Timeout(time.Second*5),
					// consul.TCPCheck(time.Second),
				)),
				server.WrapHandler(opencensus.NewHandlerWrapper()),
				server.WrapHandler(microLimiter.NewHandlerWrapper(
					ratelimit.NewBucketWithRate(5000, 5000), false),
				),
			)),
			micro.Client(client.NewClient(
				client.Selector(selector.NewSelector(selectorReg.TTL(time.Hour*2))),
				client.PoolSize(50),
				client.PoolTTL(time.Minute*5),
				client.DialTimeout(time.Second*2),
				client.RequestTimeout(time.Second*5),
			)),
			//micro.Broker(),
		)
		a.rpcService.Init()
		err := proto.RegisterGreeterHandler(a.rpcService.Server(), rpc.NewRpcHandler(a.useCase))
		if err != nil {
			return err
		}

		logger.Info("New rpc service successfully.")
		return nil
	}
}

// Handler ...
func Handler() Option {
	return func(a *app) (err error) {
		a.restHandler = rest.NewRestHandler(a.useCase)
		if a.restHandler == nil {
			return errors.New("create rest handler failed")
		}

		logger.Info("New rest handler successfully.")
		return
	}
}

// Router ...
func Router() Option {
	return func(a *app) (err error) {
		ginMode := gin.Mode()
		if ginMode == gin.ReleaseMode {
			a.ginRouter = gin.New()
			a.ginRouter.Use(gin.Recovery(), cors.Default())
		} else {
			a.ginRouter = gin.Default()
		}

		a.ginRouter.Use(middleware.NewRateLimiter(time.Second, 5000))
		rest.RegisterHandler(a.ginRouter, a.restHandler)

		a.ginRouter.NoRoute(func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusNotFound)
		})

		logger.Info("New web router successfully.")
		return
	}
}

func WebService() Option {
	return func(a *app) error {
		a.webService = web.NewService(
			//web.MicroService(micro.NewService(micro.Server(
			//	server.NewServer()))),
			web.Name("greeterRest"),
			web.Id(fmt.Sprintf("greeterRest-%v", 9999)),
			web.Version("latest"),
			web.Metadata(map[string]string{
				"nodeId":      "9999",
				"serviceName": "greeter",
				"type":        "greeterRest",
			}),
			web.RegisterTTL(time.Second*30),
			web.RegisterInterval(time.Second*10),
			web.Registry(consul.NewRegistry(
				registry.Addrs("127.0.0.1:8500"),
				registry.Timeout(time.Second*10),
			)),
			web.Address(":8086"),
			web.Flags(&cli.BoolFlag{
				Name:  "run_client",
				Usage: "Launch the client",
			}),
			web.Handler(a.ginRouter),
		)

		logger.Info("New web service successfully.")
		return nil
	}
}

func UseCase() Option {
	return func(a *app) error {
		a.useCase = service.NewUseCase(a.dao)
		return nil
	}
}

// RedisCli ...
func RedisCli() Option {
	return func(a *app) error {
		makeClientFunc := redisClient.NewClient
		//if !a.conf.IsDebugMode() {
		//	makeClientFunc = client.NewClientNewCluster
		//}

		a.redisCli = makeClientFunc(redisClient.Config{
			Address:  "127.0.0.1:6379",
			Password: "",
			PoolConfig: redis.PoolConfig{
				MaxIdle:     10,
				MaxActive:   200,
				IdleTimeout: time.Minute,
				Wait:        true,
			},
		})

		if a.redisCli == nil {
			return errors.New("create redis client failed")
		}

		logger.Info("New Redis client successfully.")
		return nil
	}
}

func MySQLCli() Option {
	return func(a *app) (err error) {
		a.mysqlCli, err = mysql.NewMysqlPoolWithTrace(&mysql.Config{
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "Admin123",
			DBName:   "db_player",
			CharSet:  "utf8mb4",
			MaxConn:  200,
			IdleConn: 10,
		})

		if err != nil {
			return err
		}

		if a.mysqlCli == nil {
			err = errors.New("create mysql client failed")
			return
		}

		logger.Info("New MySQL client successfully.")
		return
	}
}

func MongoCli() Option {
	return func(a *app) (err error) {
		a.mongoCli, err = mongo.NewClient(&mongo.Config{
			Hosts:       []string{"127.0.0.1:27017"},
			Database:    "ffa",
			UserName:    "",
			Password:    "",
			MaxPoolSize: 200,
			MinPoolSize: 10,
			MaxIdleTime: 3600,
		})

		if err != nil {
			return err
		}

		if a.mongoCli == nil {
			err = errors.New("create mongo client failed")
		}

		logger.Info("New Mongodb client successfully.")
		return
	}
}

// Dao ...
func Dao() Option {
	return func(a *app) (err error) {
		a.dao = store.NewDao(a.mysqlCli, a.mongoCli)
		if a.dao == nil {
			return errors.New("create dao failed")
		}

		logger.Info("New dao successfully.")
		return
	}
}
