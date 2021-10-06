### 资料信息
http://www.topgoer.com/  
https://github.com/asim/go-micro/tree/master/cmd/protoc-gen-micro
https://github.com/asim/go-micro/blob/master/examples/README.md
https://github.com/asim/go-micro/blob/master/examples/greeter/README.md
https://github.com/asim/go-micro/blob/master/cmd/gomu/README.md

### mDNs
mDNS 一般作用在一个局域网内的，有特定的 IP 地址，也就是 224.0.0.251,有特定的端口 5353,mDNS 的作用是实现局域网内的服务发现，查询，注册


```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get github.com/asim/go-micro/cmd/protoc-gen-micro/v3

protoc --proto_path=. --micro_out=../ --go_out=../  *.proto

go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
```

#### 优化方向
-   consul 节点ID生成、服务注册与发现、配置中心
-   go-micro 请求的熔断、响应的限流、请求的负载均衡、消息队列
-   统一的存储层 mysql、mongodb
-   缓存及分布式锁 redis、bigcache
-   APM Application Performance Management） 监控（metrics）、日志（logs）、追踪（tracing）
-   监控（Prometheus、Grafana） 操作系统、进程状态、请求信息、中间件、存储信息
-   日志 EFK 展示面板、日志报警
-   调用链追踪 zipkin
-   渐进的基础框架库 (infra) redis、mysql、mongodb、middleware、plugin、 msg protocol、http client、http server、grpc server、logger等
-   灰度发布 请求染色

#### swagger
文档 https://github.com/swaggo/swag/blob/master/README_zh-CN.md  
```shell
go get -u github.com/swaggo/swag/cmd/swag
swag init -g handler.go -d ./internal/api/rest -o ./internal/api/rest/docs --parseInternal  --generatedTime
http://localhost:8086/swagger/index.html
```

