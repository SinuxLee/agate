package main

import (
	"context"
	"fmt"
	"time"

	"template/pkg/proto"

	"github.com/asim/go-micro/plugins/client/http/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	selectorReg "github.com/asim/go-micro/plugins/selector/registry"
	"github.com/asim/go-micro/plugins/selector/shard/v3"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	"github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/gin-gonic/gin"
)

var (
	emptyData = struct{}{}
)

const (
	consulAddr = "127.0.0.1:8500"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HelloRsp struct {
	Response
	Data struct {
		Greet string `json:"greet"`
	} `json:"data,omitempty"`
}

func main() {
	rpcCli()
	webCli()
}

func webCli() {
	reg := consul.NewRegistry(
		registry.Addrs(consulAddr),
		registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selectorReg.TTL(time.Hour),
		selector.Registry(reg),
	)

	cli := http.NewClient(
		client.PoolSize(500),
		client.PoolTTL(time.Minute),
		client.Selector(sel),
		client.Retries(3),
		client.DialTimeout(time.Second*2),
		client.RequestTimeout(time.Second*5),
		client.Wrap(hystrix.NewClientWrapper()),
		client.Wrap(opencensus.NewClientWrapper()),
		client.ContentType(gin.MIMEJSON),
	)

	// 只能调用POST 方法
	rsp := &HelloRsp{}
	req := cli.NewRequest("svrWEB", "/svr/v1/hello", emptyData)
	err := cli.Call(context.TODO(), req, rsp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rsp.Data.Greet)
}

func rpcCli() {
	reg := consul.NewRegistry(
		registry.Addrs(consulAddr),
		registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selectorReg.TTL(time.Hour*2),
		selector.Registry(reg),
	)

	cli := client.NewClient(
		client.PoolSize(500),
		client.PoolTTL(time.Minute),
		client.Selector(sel),
		client.Transport(grpc.NewTransport()),
		client.Retries(3),
		client.DialTimeout(time.Second*2),
		client.RequestTimeout(time.Second*5),
		client.Wrap(hystrix.NewClientWrapper()),
		client.Wrap(opencensus.NewClientWrapper()),
	)

	// Use the generated client stub
	cl := proto.NewGreeterService("svrRPC", cli) // consul中注册的服务名称

	// Make request
	req := &proto.HelloRequest{
		Name: "John",
	}
	rsp, err := cl.Hello(context.Background(), req, shard.Strategy(req.Name)) // chash
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)
}
