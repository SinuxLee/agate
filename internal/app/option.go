package app

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"template/pkg/infra/monitoring"
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

	zlog "github.com/asim/go-micro/plugins/logger/zerolog/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	microLimiter "github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/plugins/wrapper/validator/v3"
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
	"github.com/felixge/fgprof"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/imdario/mergo"
	"github.com/juju/ratelimit"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	//	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
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
			return errors.Wrap(err, "Config()")
		}

		err = a.conf.Load(flagSource, env.NewSource())
		if err != nil {
			return err
		}

		return nil
	}
}

func Version(versionInfo string) Option {
	return func(a *app) (err error) {
		printVersion := a.conf.Get(printVersionKey).Bool(printVersionDef)
		if printVersion {
			_, _ = fmt.Fprintf(os.Stderr, "%v build info: %v\n", serverName, versionInfo)
			os.Exit(0)
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
			return errors.Wrapf(err, "get consul addr: %s", consulAddrKey)
		}

		a.nodeID, err = nodeNamed.GetNodeID(&nid.NameHolder{
			LocalPath:  os.Args[0],
			LocalIP:    ip,
			ServiceKey: serverName,
		})

		if err != nil {
			return errors.Wrapf(err, "get nodeid: %s", consulAddrKey)
		}

		return nil
	}
}

// Logger ...
func Logger() Option {
	return func(a *app) error {
		lv := a.conf.Get(logLevelKey).String(logLevelDef)
		level, err := zerolog.ParseLevel(lv)
		if err != nil {
			level = zerolog.DebugLevel
		}

		zerolog.MessageFieldName = "msg"
		zerolog.LevelFieldName = "lvl"
		zerolog.TimeFieldFormat = timeFormat
		//zerolog.InterfaceMarshalFunc = func(v interface{}) ([]byte, error) {
		//	buffer := bytes.NewBuffer([]byte{})
		//	en := json.NewEncoder(buffer)
		//	en.SetEscapeHTML(false)
		//	en.SetIndent("", "")
		//	err := en.Encode(v)
		//	return buffer.Bytes(), err
		//}

		simpleHook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			if _, file, line, ok := runtime.Caller(4); ok {
				// 取文件名
				idx := strings.LastIndexByte(file, '/')
				if idx == -1 {
					e.Str("file", fmt.Sprintf("%s:%d", file, line))
					return
				}

				// 取包名
				idx = strings.LastIndexByte(file[:idx], '/')
				if idx == -1 {
					e.Str("file", fmt.Sprintf("%s:%d", file[:idx], line))
					return
				}

				// 返回包名和文件名
				e.Str("file", fmt.Sprintf("%s:%d", file[idx+1:], line))
			}
		})

		ip, _ := a.intranetIP()
		log.Logger = zerolog.New(os.Stdout).Level(level).Hook(simpleHook).With().Timestamp().
			Fields(map[string]interface{}{"nodeId": a.nodeID}).IPAddr("ip", net.ParseIP(ip)).Logger()
		log.Info().Msg("Init logger successfully.")

		logger.DefaultLogger = zlog.NewLogger(
			logger.WithOutput(os.Stdout),
			logger.WithLevel(logger.DebugLevel),
			zlog.WithTimeFormat(timeFormat),
			zlog.WithProductionMode(),
			zlog.WithHooks([]zerolog.Hook{simpleHook}),
			logger.WithFields(map[string]interface{}{"nodeId": a.nodeID, "ip": ip}),
		)
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
			return errors.Wrapf(err, "KVStore: %s", consulAddrKey)
		}
		return nil
	}
}

// RedisCli ...
func RedisCli() Option {
	return func(a *app) error {
		conf := &redisConf{}
		err := a.getConsulConf("redis", conf, &redisConf{
			Mode:     "standalone",
			Addr:     "127.0.0.1:6379",
			Password: "",
		})
		if err != nil {
			return errors.Wrapf(err, "RedisCli(): %s", "redis")
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

		log.Info().Msg("New Redis client successfully.")
		return nil
	}
}

// MySQLCli ...
func MySQLCli() Option {
	return func(a *app) error {
		conf := &mysqlConf{}
		err := a.getConsulConf("mysql", conf, &mysqlConf{
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "Admin123",
			Database: "db_player",
		})
		if err != nil {
			return errors.Wrapf(err, "MySQLCli(): %s", "mysql")
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

		log.Info().Msg("New MySQL client successfully.")
		return nil
	}
}

// MongoCli ...
func MongoCli() Option {
	return func(a *app) (err error) {
		conf := &mongodbConf{}
		err = a.getConsulConf("mongodb", conf, &mongodbConf{
			Host:     []string{"127.0.0.1:27017"},
			User:     "",
			Password: "",
			Database: "ffa",
		})
		if err != nil {
			return errors.Wrapf(err, "MongoCli(): %s", "mongodb")
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

		log.Info().Msg("New Mongodb client successfully.")
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

		log.Info().Msg("New dao successfully.")
		return
	}
}

// UseCase ...
func UseCase() Option {
	return func(a *app) error {
		conf := &bizConf{}
		defaultCfg := &bizConf{
			ActiveBeginTime: time.Now().Format("2021-10-06 00:00:00"),
		}
		err := a.getConsulConf(bizConfKey, conf, defaultCfg)
		if err != nil && err != libKVStore.ErrKeyNotFound {
			return errors.Wrapf(err, "UseCase(): %s", bizConfKey)
		}
		mergo.Merge(conf, defaultCfg)
		a.useCase = service.NewUseCase(a.dao, conf)
		return a.watchConsulConf(bizConfKey, conf)
	}
}

// RpcService ...
func RpcService() Option {
	return func(a *app) error {
		conf := &rpcConf{}
		err := a.getConsulConf("rpc", conf, &rpcConf{
			Port: ":18086",
		})
		if err != nil {
			return errors.Wrapf(err, "RpcService(): %s", "rpc")
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
				server.WrapHandler(monitoring.GoMicroHandlerWrapper()),
				server.WrapHandler(validator.NewHandlerWrapper()),
				server.WrapHandler(opencensus.NewHandlerWrapper()),
				server.WrapHandler(microLimiter.NewHandlerWrapper(
					ratelimit.NewBucketWithQuantum(time.Second, 10000, 10000), true),
				),
			)),
		)

		err = proto.RegisterGreeterHandler(a.rpcService.Server(), rpc.NewRpcHandler(a.useCase))
		if err != nil {
			return errors.Wrapf(err, "RpcService(): %s", "rpc")
		}

		log.Info().Msg("New rpc service successfully.")
		return nil
	}
}

// WebService ...
func WebService() Option {
	return func(a *app) error {
		conf := &webConf{}
		err := a.getConsulConf("web", conf, &webConf{
			GinMode: "debug",
			Port:    ":8086",
		})
		if err != nil {
			return errors.Wrapf(err, "WebService(): %s", "web")
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
		ginRouter.Use(monitoring.GinHandler())
		ginRouter.NoRoute(func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusNotFound)
		})

		// analyze On-CPU as well as Off-CPU time
		ginRouter.GET("/debug/fgprof", gin.WrapH(fgprof.Handler()))

		// pprof
		pprof.Register(ginRouter)

		// 配置 swagger address
		ip, err := a.intranetIP()
		if err != nil {
			return errors.Wrapf(err, "WebService(): %s", "web")
		}
		swaggerAddr := fmt.Sprintf("%v%v", ip, conf.Port)

		// http prometheus
		//ginProm := ginPrometheus.NewPrometheus(serverName)

		// 构建 web handler
		rest.NewRestHandler(a.useCase, swaggerAddr, nil).RegisterHandler(ginRouter)

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

		log.Info().Msg("New web service successfully.")
		return nil
	}
}

func Monitor() Option {
	return func(a *app) error {
		conf := &monitoring.Config{}
		defaultCfg := &monitoring.Config{
			Addr:       ":9100",
			Path:       "/metrics",
			ServerName: serverName,
		}

		err := a.getConsulConf("metrics", conf, defaultCfg)

		if err != nil && err != libKVStore.ErrKeyNotFound {
			return errors.Wrapf(err, "Monitor(): %s", "metrics")
		}

		mergo.Merge(conf, defaultCfg)
		err = monitoring.Serve(conf)
		return errors.Wrapf(err, "Monitor(): %s", "metrics")
	}
}
