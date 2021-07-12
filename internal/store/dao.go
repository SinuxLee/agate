package store

import (
	"context"
	"template/pkg/infra/mongo"
	"template/pkg/infra/mysql"
)

var _ Dao = (*daoImpl)(nil)

type Dao interface {
	Hello(ctx context.Context, name string) (string, error)
}

func NewDao(mysqlCli mysql.Client, mongoCli mongo.Client) Dao {
	return &daoImpl{
		sqlRepo:   mysqlCli,
		mongoRepo: mongoCli,
	}
}

type daoImpl struct {
	sqlRepo   mysql.Client
	mongoRepo mongo.Client
}
