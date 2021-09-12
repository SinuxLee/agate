module template

go 1.14

require (
	contrib.go.opencensus.io/integrations/ocsql v0.1.7
	github.com/asim/go-micro/plugins/client/http/v3 v3.0.0-20210726052521-c3107e6843e2
	github.com/asim/go-micro/plugins/logger/zerolog/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/registry/consul/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/selector/registry v0.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/selector/shard/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/transport/grpc/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3 v3.0.0-20210709115208-3fbf2c304fe0
	github.com/asim/go-micro/v3 v3.5.2
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/bwmarrin/snowflake v0.3.0
	github.com/docker/libkv v0.2.1
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gomodule/redigo v1.8.5
	github.com/jmoiron/sqlx v1.3.4
	github.com/juju/ratelimit v1.0.1
	github.com/julianshen/gin-limiter v0.0.0-20161123033831-fc39b5e90fe7
	github.com/mna/redisc v1.2.1
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/xid v1.2.1
	github.com/rs/zerolog v1.23.0
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	go.mongodb.org/mongo-driver v1.5.4
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/protobuf v1.26.0
)

replace github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.9.1
