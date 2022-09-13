package rpc

import (
	"context"

	"template/pkg/proto"

	"github.com/asim/go-micro/v3/metadata"
	"github.com/rs/zerolog/log"
)

func (r *rpcHandler) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	md, _ := metadata.FromContext(ctx)
	log.Info().Interface("metadata", md).Msg("receive request")
	data, err := r.useCase.Hello(ctx, req.Name)
	if err != nil {
		log.Err(err).Msg("bad request")
		return err
	}

	rsp.Greeting = data
	return nil
}
