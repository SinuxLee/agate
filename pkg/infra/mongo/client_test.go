package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var conf = &Config{
	Hosts:       []string{"localhost:27017"},
	MaxPoolSize: 100,
	MinPoolSize: 100,
	MaxIdleTime: 60,
	UserName:    "test",
	Password:    "test",
	Database:    "test",
}

type BsonData struct {
	Name string `bson:"name" json:"name"`
}

func TestNewClient(t *testing.T) {
	cli, err := NewClient(conf)
	if err != nil {
		t.Fatal(err.Error())
	}
	data := &BsonData{}
	if err := cli.FindOne("libz", bson.M{"name": "libz"}, data); err != nil {
		t.Fatal("failed to find")
	}
}
