package mongo

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type TraverseFunc func(data interface{}) error

var _ Client = (*client)(nil)

type Client interface {
	FindOne(ctx context.Context, table string, finder interface{}, data interface{}) error
	Find(ctx context.Context, table string, finder interface{}, data interface{}) error
	UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}) error
	MultiReplaceInsert(ctx context.Context, table string, filter []interface{}, data []interface{}) error
	RunJavascript(ctx context.Context, script string) ([]interface{}, error)
	Traverse(ctx context.Context, table string, finder interface{}, data interface{}, projection interface{}, limit int64, fun TraverseFunc) error
	Transaction(ctx context.Context, table string) error
	Session(ctx context.Context, table string) error
}

type Config struct {
	Hosts       []string
	Database    string
	UserName    string
	Password    string
	MaxPoolSize uint
	MinPoolSize uint
	MaxIdleTime uint
}

func NewClient(conf *Config) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	opts := options.Client().SetMaxPoolSize(uint64(conf.MaxPoolSize)).SetMinPoolSize(uint64(conf.MinPoolSize)).
		SetMaxConnIdleTime(time.Duration(conf.MaxIdleTime) * time.Second).SetHosts(conf.Hosts)
	if conf.UserName != "" && conf.Password != "" {
		opts = opts.SetAuth(options.Credential{
			Username:   conf.UserName,
			Password:   conf.Password,
			AuthSource: conf.Database,
		})
	}

	cli, err := mongo.Connect(ctx, opts)
	defer cancel()

	if err != nil {
		return nil, err
	}

	// 判断服务是否可用
	if err = cli.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, err
	}

	return &client{
		cli:  cli,
		conf: conf,
	}, nil
}

type client struct {
	cli  *mongo.Client
	conf *Config
}

func (c *client) FindOne(ctx context.Context, table string, finder interface{}, data interface{}) error {
	// fixme 没必要每次都创建Collection
	collection := c.cli.Database(c.conf.Database).Collection(table)
	return collection.FindOne(ctx, finder).Decode(data)
}

func (c *client) Find(ctx context.Context, table string, finder interface{}, data interface{}) error {
	collection := c.cli.Database(c.conf.Database).Collection(table)
	cursor, err := collection.Find(ctx, finder)
	if err != nil {
		return err
	}

	return cursor.All(context.TODO(), data)
}

func (c *client) UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}) error {
	collection := c.cli.Database(c.conf.Database).Collection(table)
	_, err := collection.UpdateOne(ctx, filter, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) Traverse(ctx context.Context, table string, finder interface{}, data interface{}, projection interface{}, limit int64, fun TraverseFunc) error {
	collection := c.cli.Database(c.conf.Database).Collection(table)
	batchSize := int32(200)
	skipCount := int64(0)

	cursor, err := collection.Find(context.TODO(), finder, &options.FindOptions{
		BatchSize:  &batchSize,
		Skip:       &skipCount,
		Projection: projection,
		Limit:      &limit,
	})

	if err != nil {
		return err
	}

	for cursor.TryNext(context.TODO()) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = cursor.Decode(data)
			if err != nil {
				return err
			}

			err := fun(data)
			if err != nil {
				return err
			}

			length := cursor.RemainingBatchLength()
			if length == 0 {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}

	return nil
}

// MultiReplaceInsert ...
func (c *client) MultiReplaceInsert(ctx context.Context, table string, filter []interface{}, data []interface{}) error {
	length := len(filter)
	if length != len(data) {
		return errors.New("filter and data do not match")
	}

	models := make([]mongo.WriteModel, 0, length)
	for i := 0; i < length; i++ {
		model := mongo.NewReplaceOneModel().SetFilter(filter[i])
		model.SetReplacement(data[i])
		model.SetUpsert(true)
		models = append(models, model)
	}

	collection := c.cli.Database(c.conf.Database).Collection(table)
	_, err := collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return err
	}

	return nil
}

// RunJavascript 需要mongodb 4.0之前的版本
func (c *client) RunJavascript(ctx context.Context, script string) ([]interface{}, error) {
	js := bsonx.JavaScript(script)
	res, err := c.cli.Database(c.conf.Database).RunCommand(ctx, bson.M{"eval": js}).DecodeBytes()
	if err != nil {
		return nil, err
	}

	items, err := res.LookupErr("retval")
	if err != nil {
		return nil, err
	}

	arr := make([]interface{}, 0)
	err = items.Unmarshal(&arr)
	if err != nil {
		return nil, err
	}

	return arr, nil
}

// Transaction ...
func (c *client) Transaction(ctx context.Context, table string) error {
	session, err := c.cli.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// session 上下文来保证一个事务
	sessCtx := mongo.NewSessionContext(ctx, session)

	//开始事务
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// 可以用多个表，保证事务执行
	collection := session.Client().Database(c.conf.Database).Collection(table)

	_, err = collection.InsertOne(sessCtx, bson.M{"_id": "222", "name": "ddd", "age": 50})
	if err != nil {
		return err
	}

	//写重复id
	_, err = collection.InsertOne(sessCtx, bson.M{"_id": "111", "name": "ddd", "age": 50})
	if err != nil {
		_ = session.AbortTransaction(ctx)
		return err
	}

	return session.CommitTransaction(ctx)
}

// Session 需要Replication Set才能执行
func (c *client) Session(ctx context.Context, table string) error {
	return c.cli.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		col := sessionContext.Client().Database(c.conf.Database).Collection(table)
		_, err = col.InsertOne(sessionContext, bson.M{"_id": "444", "name": "ddd", "age": 50})
		if err != nil {
			return err
		}

		_, err = col.InsertOne(sessionContext, bson.M{"_id": "111", "name": "ddd", "age": 50})
		if err != nil {
			_ = sessionContext.AbortTransaction(sessionContext)
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})
}
