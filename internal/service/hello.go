package service

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

func (uc *useCaseImpl) Hello(ctx context.Context, name string) (string, error) {
	data, err := uc.dao.Hello(ctx, name)
	if err != nil {
		return "", err
	}

	log.Info().Str("thirdParty", uc.conf.GetThirdParty()).Send()

	if rsp, err := resty.New().SetHostURL(uc.conf.GetThirdParty()).R().Get("/anything/haha"); err == nil && rsp.IsSuccess() {
		log.Info().Str("body", rsp.String()).Msg("response")
	}

	return data, nil
}
