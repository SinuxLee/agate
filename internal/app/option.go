package app

import (
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
	"template/pkg/infra/nid"
	"template/pkg/middleware"
	"template/pkg/proto"

	"github.com/asim/go-micro/plugins/logger/zerolog/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	microLimiter "github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/config/source/env"
	"github.com/asim/go-micro/v3/config/source/flag"
	"github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/web"
	"github.com/docker/libkv"
	libKVStore "github.com/docker/libkv/store"
	libKVConsul "github.com/docker/libkv/store/consul"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/juju/ratelimit"
	"github.com/pkg/errors"
)

// Option ...
type Option func(*app) error

// Config ..
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

		return nil
	}
}

// NodeID ...
func NodeID() Option {
	return func(a *app) error {
		ip, err := a.intranetIP()
		if err != nil {
			return err
		}

		addr := a.conf.Get(consulAddrKey).String(consulAddrDef)
		nodeNamed, err := nid.NewConsulNamed(addr)
		if err != nil {
			return err
		}

		a.nodeID, err = nodeNamed.GetNodeID(&nid.NameHolder{
			LocalPath:  os.Args[0],
			LocalIP:    ip,
			ServiceKey: serverName,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

// Logger ...
func Logger() Option {
	return func(a *app) error {
		lv := a.conf.Get(logLevelKey).String(logLevelDef)
		level, err := logger.GetLevel(lv)
		if err != nil {
			level = logger.DebugLevel
		}

		logger.DefaultLogger = zerolog.NewLogger(logger.WithOutput(os.Stdout),
			logger.WithLevel(level), logger.WithFields(map[string]interface{}{"nodeId": a.nodeID}))
		logger.Info("Init logger successfully.")
		return nil
	}
}

// KVStore ...
func KVStore() Option {
	return func(a *app) (err error) {
		libKVConsul.Register()
		addr := a.conf.Get(consulAddrKey).String(consulAddrDef)
		a.kvStore, err = libkv.NewStore(libKVStore.CONSUL, []string{addr},
			&libKVStore.Config{ConnectionTimeout: 10 * time.Second})
		if err != nil {
			return err
		}
		return nil
	}
}

// RedisCli ...
func RedisCli() Option {
	return func(a *app) error {
		conf := &redisConf{}
		err := a.getConsulConf("redis", conf)
		if err != nil {
			return err
		}

		if conf.Mode == redisModeCluster {
			a.redisCli = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:        []string{conf.Addr},
				MaxRedirects: 3,
				Password:     conf.Password,
				MaxRetries:   3,
				PoolSize:     200,
				MinIdleConns: 20,
				MaxConnAge:   time.Hour,
				IdleTimeout:  time.Minute,
			})
		} else {
			a.redisCli = redis.NewClient(&redis.Options{
				Addr:         conf.Addr,
				Password:     conf.Password,
				MaxRetries:   3,
				PoolSize:     200,
				MinIdleConns: 20,
				MaxConnAge:   time.Hour,
				IdleTimeout:  time.Minute,
			})
		}

		if a.redisCli == nil {
			return errors.Errorf("create redis client failed, %v", conf.Info())
		}

		logger.Info("New Redis client successfully.")
		return nil
	}
}

// MySQLCli ...
func MySQLCli() Option {
	return func(a *app) error {
		conf := &mysqlConf{}
		err := a.getConsulConf("mysql", conf)
		if err != nil {
			return err
		}

		a.mysqlCli, err = mysql.NewMysqlPoolWithTrace(&mysql.Config{
			Host:     conf.Host,
			Port:     conf.Port,
			User:     conf.User,
			Password: conf.Password,
			DBName:   conf.Database,
			CharSet:  "utf8mb4",
			MaxConn:  200,
			IdleConn: 10,
		})

		if err != nil {
			return errors.Wrapf(err, "%v", conf.Info())
		}

		if a.mysqlCli == nil {
			return errors.Errorf("create mysql client failed, %v", conf.Info())
		}

		logger.Info("New MySQL client successfully.")
		return nil
	}
}

// MongoCli ...
func MongoCli() Option {
	return func(a *app) (err error) {
		conf := &mongodbConf{}
		err = a.getConsulConf("mongodb", conf)
		if err != nil {
			return err
		}

		a.mongoCli, err = mongo.NewClient(&mongo.Config{
			Hosts:       conf.Host,
			Database:    conf.Database,
			UserName:    conf.User,
			Password:    conf.Password,
			MaxPoolSize: 200,
			MinPoolSize: 10,
			MaxIdleTime: 3600,
		})

		if err != nil {
			return errors.Wrapf(err, "%v", conf.Info())
		}

		if a.mongoCli == nil {
			err = errors.Errorf("create mongo client failed, %v", conf.Info())
		}

		logger.Info("New Mongodb client successfully.")
		return
	}
}

// Dao ...
func Dao() Option {
	return func(a *app) (err error) {
		a.dao = store.NewDao(a.redisCli, a.mysqlCli, a.mongoCli)
		if a.dao == nil {
			return errors.New("create dao failed")
		}

		logger.Info("New dao successfully.")
		return
	}
}

// UseCase ...
func UseCase() Option {
	return func(a *app) error {
		a.useCase = service.NewUseCase(a.dao)
		return nil
	}
}

// RpcService ...
func RpcService() Option {
	return func(a *app) error {
		conf := &rpcConf{}
		err := a.getConsulConf("rpc", conf)
		if err != nil {
			return err
		}

		consulAddr := a.conf.Get(consulAddrKey).String(consulAddrDef)
		srvName := fmt.Sprintf("%vRPC", serverName)
		serverID := fmt.Sprintf("%02v", a.nodeID)
		a.rpcService = micro.NewService(
			micro.Server(server.NewServer(
				server.Name(srvName), // consul 中的 service name
				server.Id(serverID),
				server.Transport(grpc.NewTransport()),
				server.Metadata(map[string]string{
					"nodeId":      serverID,
					"serviceName": srvName,
					"type":        "rpc",
				}),
				server.Address(conf.Port),
				server.Registry(consul.NewRegistry(
					registry.Addrs(consulAddr),
				)),
				server.WrapHandler(opencensus.NewHandlerWrapper()),
				server.WrapHandler(microLimiter.NewHandlerWrapper(
					ratelimit.NewBucketWithQuantum(time.Second, 10000, 10000), false),
				),
			)),
		)

		err = proto.RegisterGreeterHandler(a.rpcService.Server(), rpc.NewRpcHandler(a.useCase))
		if err != nil {
			return err
		}

		logger.Info("New rpc service successfully.")
		return nil
	}
}

// WebService ...
func WebService() Option {
	return func(a *app) error {
		conf := &webConf{}
		err := a.getConsulConf("web", conf)
		if err != nil {
			return err
		}

		// consul配置优先级高于环境变量
		if conf.GinMode != gin.DebugMode {
			gin.SetMode(conf.GinMode)
		}

		// 创建路由
		var ginRouter *gin.Engine
		ginMode := gin.Mode()
		if ginMode != gin.DebugMode {
			ginRouter = gin.New()
			if ginMode == gin.TestMode {
				ginRouter.Use(gin.Logger())
			}
			ginRouter.Use(gin.Recovery())
		} else {
			ginRouter = gin.Default()
		}
		ginRouter.Use(cors.Default(), middleware.NewRateLimiter(time.Second, 10000))
		ginRouter.NoRoute(func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusNotFound)
		})
		pprof.Register(ginRouter)

		// 配置 swagger address
		ip, err := a.intranetIP()
		if err != nil {
			return err
		}
		swaggerAddr := fmt.Sprintf("%v%v", ip, conf.Port)

		// 构建 web handler
		rest.NewRestHandler(a.useCase, swaggerAddr).RegisterHandler(ginRouter)

		// 注册服务
		consulAddr := a.conf.Get(consulAddrKey).String(consulAddrDef)
		webName := fmt.Sprintf("%vWEB", serverName)
		webID := fmt.Sprintf("%v-%02v", webName, a.nodeID)
		a.webService = web.NewService(
			web.Name(webName),
			web.Id(webID),
			web.Metadata(map[string]string{
				"nodeId":      webID,
				"serviceName": webName,
				"type":        "web",
				"protocol":    "http",
			}),
			web.Registry(consul.NewRegistry(
				registry.Addrs(consulAddr),
			)),
			web.Address(conf.Port),
			web.Handler(ginRouter),
		)

		logger.Info("New web service successfully.")
		return nil
	}
}
