package snowflake

import innerSnowflake "github.com/bwmarrin/snowflake"

var node *innerSnowflake.Node

func InitSnowflake(id int) (err error) {
	node, err = innerSnowflake.NewNode(int64(id))
	return
}

func GenerateInt64() int64 {
	return node.Generate().Int64()
}

func GenerateBase32() string {
	return node.Generate().Base32()
}
