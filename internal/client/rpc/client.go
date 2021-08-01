package rpc

import (
	"fmt"
	"time"

	"template/pkg/proto"

	"github.com/asim/go-micro/plugins/registry/consul/v3"
	selectorReg "github.com/asim/go-micro/plugins/selector/registry"
	"github.com/asim/go-micro/plugins/transport/grpc/v3"
	"github.com/asim/go-micro/plugins/wrapper/breaker/hystrix/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opencensus/v3"
	microClient "github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
)

const (
	svrService = "svr"
)

type GreeterClient interface {
	proto.GreeterService
}

func NewGreeterClient(consulAddr string) GreeterClient {
	reg := consul.NewRegistry(
		registry.Addrs(consulAddr),
		registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selectorReg.TTL(time.Hour),
		selector.Registry(reg),
	)

	cli := microClient.NewClient(
		microClient.PoolSize(500),
		microClient.PoolTTL(time.Minute),
		microClient.Selector(sel),
		microClient.Transport(grpc.NewTransport()),
		microClient.Retries(3),
		microClient.DialTimeout(time.Second*2),
		microClient.RequestTimeout(time.Second*5),
		microClient.Wrap(hystrix.NewClientWrapper()),
		microClient.Wrap(opencensus.NewClientWrapper()),
	)

	return proto.NewGreeterService(fmt.Sprintf("%vRPC", svrService), cli) // consul中注册的服务名称
}
