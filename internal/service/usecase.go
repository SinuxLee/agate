package service

import (
	"context"
	"template/internal/store"
)

var _ UseCase = (*useCaseImpl)(nil)

type UseCase interface {
	Hello(ctx context.Context, name string) (string, error)
}

func NewUseCase(d store.Dao) UseCase {
	return &useCaseImpl{
		dao: d,
	}
}

type useCaseImpl struct {
	dao store.Dao
}
