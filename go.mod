module template

go 1.14

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/asim/go-micro/plugins/client/http/v3 v3.0.0-20210726052521-c3107e6843e2
	github.com/asim/go-micro/plugins/logger/zerolog/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/registry/consul/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/selector/registry v0.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/selector/shard/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/transport/grpc/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3 v3.0.0-20211006165514-a99a1e935651
	github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/validator/v3 v3.0.0-20211002121322-2ef523a7eb0c
	github.com/asim/go-micro/v3 v3.6.1-0.20210831082736-088ccb50019c
	github.com/bwmarrin/snowflake v0.3.0
	github.com/docker/libkv v0.2.1
	github.com/felixge/fgprof v0.9.1
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.7.4
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-resty/resty/v2 v2.1.1-0.20191201195748-d7b97669fe48
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/gops v0.3.20
	github.com/hedemonde/go-gin-prometheus v0.1.2
	github.com/imdario/mergo v0.3.12
	github.com/jmoiron/sqlx v1.3.4
	github.com/juju/ratelimit v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.23.0
	github.com/sinuxlee/gin-limiter v1.0.1
	github.com/swaggo/gin-swagger v1.3.2
	github.com/swaggo/swag v1.7.3
	github.com/valyala/bytebufferpool v1.0.0
	go.mongodb.org/mongo-driver v1.5.4
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/protobuf v1.26.0
)

replace github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.9.1
