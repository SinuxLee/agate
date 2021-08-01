package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var _ Client = (*client)(nil)

//Config mysql配置信息
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
	// Register our ocsql wrapper for the provided mysql driver.
	driverName, err := ocsql.Register("mysql", ocsql.WithOptions(ocsql.TraceOptions{
		AllowRoot:    true,
		Ping:         false,
		RowsNext:     false,
		RowsClose:    false,
		RowsAffected: false,
		LastInsertID: false,
		Query:        true,
		QueryParams:  true,
	}))
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driverName, cfg.getSource())
	if err != nil {
		return nil, err
	}

	// Wrap our *sql.DB with sqlx. use the original db driver name!!!
	pool := sqlx.NewDb(db, "mysql")
	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	// record DB connection pool statistics
	// dbStatsCloser := ocsql.RecordStats(db, 5*time.Second)
	// defer dbStatsCloser() TODO
	ocsql.RecordStats(db, 5*time.Second)

	pool.SetConnMaxLifetime(time.Hour * 4)
	pool.SetMaxOpenConns(cfg.MaxConn)
	pool.SetMaxIdleConns(cfg.IdleConn)

	return &client{db: pool}, nil
}

type client struct {
	db *sqlx.DB
}

func (c *client) QuerySingle(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.db.Get(dest, query, args...)
}

func (c *client) QueryMulti(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.db.Select(dest, query, args...)
}

func (c *client) Insert(_ context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil { //TODO TO LOG
		return 0, err
	}
	return lastInsertId, nil
}

func (c *client) InsertNamed(_ context.Context, query string, arg interface{}) (int64, error) {
	result, err := c.db.NamedExec(query, arg)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil { //TODO TO LOG
		return 0, err
	}
	return lastInsertId, nil
}

func (c *client) Update(_ context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	rowAffects, err := result.RowsAffected()
	if err != nil { //TODO TO LOG
		return 0, err
	}
	return rowAffects, nil
}

func (c *client) UpdateNamed(_ context.Context, query string, arg interface{}) (int64, error) {
	result, err := c.db.NamedExec(query, arg)
	if err != nil {
		return 0, err
	}
	rowAffects, err := result.RowsAffected()
	if err != nil { //TODO TO LOG
		return 0, err
	}
	return rowAffects, nil
}

func (c *client) Exec(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	rlt, err := c.db.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return rlt, err
}

func (c *client) IsNoRowsError(err error) bool {
	return errors.Cause(err) == sql.ErrNoRows
}

func (c *client) GetOriginalSource() interface{} {
	return c.db
}

func (c *client) ReplaceIntoMulti(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	multiQuery, multiArgs, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	result, err := c.db.Exec(multiQuery, multiArgs...)
	if err != nil {
		return nil, err
	}

	return result, nil
}
