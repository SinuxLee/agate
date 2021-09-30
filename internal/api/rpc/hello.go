package rpc

import (
	"context"

	"github.com/rs/zerolog/log"

	"template/pkg/proto"
)

func (r *rpcHandler) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	log.Info().Msg("receive request")
	data, err := r.useCase.Hello(ctx, req.Name)
	if err != nil {
		return err
	}

	rsp.Greeting = data
	return nil
}
