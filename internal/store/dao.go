package store

import (
	"context"
	"template/pkg/infra/mongo"
	"template/pkg/infra/mysql"

	"github.com/go-redis/redis/v8"
)

var _ Dao = (*daoImpl)(nil)

type Dao interface {
	Hello(ctx context.Context, name string) (string, error)
}

func NewDao(redisCli redis.UniversalClient, mysqlCli mysql.Client, mongoCli mongo.Client) Dao {
	return &daoImpl{
		redisRepo: redisCli,
		sqlRepo:   mysqlCli,
		mongoRepo: mongoCli,
	}
}

type daoImpl struct {
	redisRepo redis.UniversalClient
	sqlRepo   mysql.Client
	mongoRepo mongo.Client
}
