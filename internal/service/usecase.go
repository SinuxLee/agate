package service

import (
	"context"
	"template/internal/store"
)

var _ UseCase = (*useCaseImpl)(nil)

type config interface {
	GetThirdParty() string
}

type UseCase interface {
	Hello(ctx context.Context, name string) (string, error)
}

func NewUseCase(d store.Dao, conf config) UseCase {
	return &useCaseImpl{
		dao:  d,
		conf: conf,
	}
}

type useCaseImpl struct {
	dao  store.Dao
	conf config
}
