package service

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (uc *useCaseImpl) Hello(ctx context.Context, name string) (string, error) {
	data, err := uc.dao.Hello(ctx, name)
	if err != nil {
		return "", err
	}

	log.Info().Str("beginTime", uc.conf.GetActiveBeginTime()).Send()
	return data, nil
}
