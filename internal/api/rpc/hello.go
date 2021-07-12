package rpc

import (
	"context"

	"template/pkg/proto"
)

func (r *rpcHandler) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	data, err := r.useCase.Hello(ctx, req.Name)
	if err != nil {
		return err
	}

	rsp.Greeting = data
	return nil
}
