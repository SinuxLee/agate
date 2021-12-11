package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"template/pkg/infra/monitoring"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var _ Client = (*client)(nil)

// Config mysql配置信息
type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	DBName      string `json:"dbname"`
	CharSet     string `json:"charset"`
	MaxConn     int    `json:"maxConn"`
	IdleConn    int    `json:"idleConn"`
	IdleTimeout int    `json:"idleTimeout"`
}

func (cf *Config) getSource() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		cf.User, cf.Password, cf.Host, cf.Port, cf.DBName, cf.CharSet)
}

type Client interface {
	// QuerySingle ...
	QuerySingle(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// QueryMulti ...
	QueryMulti(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Insert 使用占位符 ? 传参  insert into tb (name, id) values (?, ?)
	Insert(ctx context.Context, query string, args ...interface{}) (int64, error)

	// InsertNamed 使用反射 传参  insert into tb (name, id) values (:name, :id) :字段名小写
	// arg 为obj 或 map[string]interface{} key为 参数名（:name）
	InsertNamed(ctx context.Context, query string, arg interface{}) (int64, error)

	// Update 使用占位符 ? 传参  update tb set name = ? where id = ?
	Update(ctx context.Context, query string, args ...interface{}) (int64, error)

	// UpdateNamed ...
	UpdateNamed(ctx context.Context, query string, arg interface{}) (int64, error)

	// Exec ...
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// IsNoRowsError ...
	IsNoRowsError(err error) bool

	// GetOriginalSource 获取内部的db源
	GetOriginalSource() interface{}

	// ReplaceIntoMulti ...
	ReplaceIntoMulti(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// NewMysqlPoolWithTrace 构建带trace的sqlpool
func NewMysqlPoolWithTrace(cfg *Config) (Client, error) {
	db, err := sql.Open("mysql", cfg.getSource())
	if err != nil {
		return nil, err
	}

	// Wrap our *sql.DB with sqlx. use the original db driver name!!!
	pool := sqlx.NewDb(db, "mysql")
	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	pool.SetConnMaxLifetime(time.Hour * 4)
	pool.SetMaxOpenConns(cfg.MaxConn)
	pool.SetMaxIdleConns(cfg.IdleConn)

	return &client{db: pool}, nil
}

type client struct {
	db *sqlx.DB
}

func (c *client) QuerySingle(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("QuerySingle", "")
	err := c.db.GetContext(ctx, dest, query, args)
	statHandler(err)

	return err
}

func (c *client) QueryMulti(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("QueryMulti", "")
	err := c.db.SelectContext(ctx, dest, query, args...)
	statHandler(err)
	return err
}

func (c *client) Insert(ctx context.Context, query string, args ...interface{}) (int64, error) {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("Insert", "")
	result, err := c.db.ExecContext(ctx, query, args...)
	statHandler(err)

	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (c *client) InsertNamed(ctx context.Context, query string, arg interface{}) (int64, error) {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("InsertNamed", "")
	result, err := c.db.NamedExecContext(ctx, query, arg)
	statHandler(err)

	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (c *client) Update(ctx context.Context, query string, args ...interface{}) (int64, error) {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("Update", "")
	result, err := c.db.ExecContext(ctx, query, args...)
	statHandler(err)

	if err != nil {
		return 0, err
	}
	rowAffects, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowAffects, nil
}

func (c *client) UpdateNamed(ctx context.Context, query string, arg interface{}) (int64, error) {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("UpdateNamed", "")
	result, err := c.db.NamedExecContext(ctx, query, arg)
	statHandler(err)

	if err != nil {
		return 0, err
	}
	rowAffects, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowAffects, nil
}

func (c *client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	statHandler := monitoring.GetRecordMysqlCallStatsHandler("Exec", "")
	result, err := c.db.ExecContext(ctx, query, args...)
	statHandler(err)
	return result, err
}

func (c *client) IsNoRowsError(err error) bool {
	return errors.Cause(err) == sql.ErrNoRows
}

func (c *client) GetOriginalSource() interface{} {
	return c.db
}

func (c *client) ReplaceIntoMulti(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	multiQuery, multiArgs, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	statHandler := monitoring.GetRecordMysqlCallStatsHandler("ReplaceIntoMulti", "")
	result, err := c.db.ExecContext(ctx, multiQuery, multiArgs...)
	statHandler(err)

	if err != nil {
		return nil, err
	}

	return result, nil
}
