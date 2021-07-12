package main

import (
	"context"
	"fmt"
	"time"

	"template/pkg/proto"

	"github.com/asim/go-micro/plugins/registry/consul/v3"
	selectorReg "github.com/asim/go-micro/plugins/selector/registry"
	"github.com/asim/go-micro/plugins/selector/shard/v3"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	"github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
)

func main() {
	reg := consul.NewRegistry(
		registry.Addrs("127.0.0.1:8500"),
		registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selectorReg.TTL(time.Hour*2),
		selector.Registry(reg),
	)

	cli := client.NewClient(
		client.PoolSize(50),
		client.PoolTTL(time.Minute*5),
		client.Selector(sel),
		client.Transport(grpc.NewTransport()),
		client.Retries(3),
		client.DialTimeout(time.Second*2),
		client.RequestTimeout(time.Second*5),
		client.Wrap(hystrix.NewClientWrapper()),
		client.Wrap(opencensus.NewClientWrapper()),
	)

	// create a new service
	service := micro.NewService()

	// parse command line flags
	service.Init()

	// Use the generated client stub
	cl := proto.NewGreeterService("greeterRpc", cli) // consul中注册的服务名称

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
