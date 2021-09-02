package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var conf = &Config{
	Hosts:       []string{"localhost:27017"},
	MaxPoolSize: 100,
	MinPoolSize: 100,
	MaxIdleTime: 60,
	UserName:    "",
	Password:    "",
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
	if err := cli.FindOne(context.Background(), "libz", bson.M{"name": "libz"}, data); err == nil {
		t.Fatal("failed to find")
	}

	insertData := bson.M{
		"$set": bson.M{"name": "ffa"},
	}

	if err = cli.UpsertOne(context.Background(), "libz", bson.M{"name": "ffa"}, insertData); err != nil {
		t.Fatal("failed to upsert")
	}

	if err := cli.FindOne(context.Background(), "libz", bson.M{"name": "ffa"}, data); err != nil {
		t.Fatal("failed to find")
	}

	if data.Name != "ffa" {
		t.Fatalf("data.Name is %s, not ffa", data.Name)
	}

	insertData = bson.M{
		"$set": bson.M{"name": "family"},
	}
	if err = cli.UpsertOne(context.Background(), "libz", bson.M{"name": "family"}, insertData); err != nil {
		t.Fatal("failed to upsert")
	}

	if err := cli.FindOne(context.Background(), "libz", bson.M{"name": "family"}, data); err != nil {
		t.Fatal("failed to find")
	}

	if data.Name != "family" {
		t.Fatalf("data.Name is %s, not family", data.Name)
	}
}
