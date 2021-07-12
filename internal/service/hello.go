package service

import (
	"context"
)

func (uc *useCaseImpl) Hello(ctx context.Context, name string) (string, error) {
	data, err := uc.dao.Hello(ctx, name)
	if err != nil {
		return "", err
	}

	return data, nil
}
