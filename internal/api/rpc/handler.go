package rpc

import (
	"template/internal/service"
	"template/pkg/proto"
)

var _ proto.GreeterHandler = (*rpcHandler)(nil)

func NewRpcHandler(uc service.UseCase) proto.GreeterHandler {
	return &rpcHandler{
		useCase: uc,
	}
}

type rpcHandler struct {
	useCase service.UseCase
}
