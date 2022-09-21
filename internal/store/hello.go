package store

import "context"

type HellResult struct {
	Data string `db:"data"`
}

func (d *daoImpl) Hello(ctx context.Context, name string) (string, error) {
	// return d.redisRepo.Get(ctx, name).Result()
	return "biubiu", nil
}
