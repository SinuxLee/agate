package http

import (
	"context"
	"fmt"
	"time"

	"github.com/asim/go-micro/plugins/client/http/v3"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	selectorReg "github.com/asim/go-micro/plugins/selector/registry"
	microClient "github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
)

var (
	emptyData = struct{}{}
)

const (
	svrService  = "svr"
	methodHello = "/svr/v1/hello"
)

type Client interface {
	Hello(ctx context.Context, rsp interface{}) error
}

func NewClient(consulAddr string) (Client, error) {
	reg := consul.NewRegistry(
		registry.Addrs(consulAddr),
		// registry.Timeout(time.Second*10),
	)

	sel := selector.NewSelector(
		selectorReg.TTL(time.Hour),
		selector.Registry(reg),
	)

	cli := http.NewClient(
		microClient.Selector(sel),
		microClient.Retries(3),
		//microClient.Wrap(hystrix.NewClientWrapper()),
		//microClient.Wrap(opencensus.NewClientWrapper()),
		microClient.ContentType("application/json"),
	)

	return &client{cli: cli}, nil
}

type client struct {
	cli microClient.Client
}

func (c *client) Hello(ctx context.Context, rsp interface{}) error {
	return c.call(ctx, svrService, methodHello, nil, rsp)
}

// Call 只能调用POST 方法
func (c *client) call(ctx context.Context, service, method string, req, rsp interface{}) error {
	if req == nil {
		req = emptyData
	}

	request := c.cli.NewRequest(fmt.Sprintf("%vWEB", service), method, req)
	err := c.cli.Call(ctx, request, rsp)
	if err != nil {
		return err
	}

	return nil
}
