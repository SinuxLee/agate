package client

import (
	"errors"
	"io/ioutil"
	"path"
	"time"

	"template/pkg/infra/redis"

	"github.com/rs/zerolog/log"
)

// lua脚本
const (
	luaExt = ".lua"

	HMGET      = "hmget.lua"
	SADD       = "sadd.lua"
	SRAND      = "srand.lua"
	STOPCAR    = "stopcar.lua"
	TAKECAR    = "takecar.lua"
	ADDCARPORT = "addcarport.lua"
	OFFLINEMSG = "offlinemsg.lua"
)

// 脚本名称和参数个数的映射
var scriptParam = map[string]int{
	HMGET: 1,
	SADD:  2,
	SRAND: 2,
	// STOPCAR:    1,
	// TAKECAR:    1,
	ADDCARPORT: 1,
	OFFLINEMSG: 1,
}
var scriptMap map[string]*redis.Script
var rdsCli RedisClient

// InitScript ...
func InitScript(cli RedisClient, path string) error {
	rdsCli = cli
	scriptMap = make(map[string]*redis.Script)
	err, files := searchFile(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if val, ok := scriptParam[file]; ok {
			err := loadLuaScript(path, file, val)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func searchFile(pathName string) (error, []string) {
	files, err := ioutil.ReadDir(pathName)
	if err != nil {
		return err, nil
	}

	result := make([]string, 0)
	for _, file := range files {
		name := file.Name()
		if path.Ext(name) == luaExt && !file.IsDir() {
			result = append(result, name)
		}
	}

	return nil, result
}

// fixme: merge to client

func loadLuaScript(path, file string, keyCount int) error {
	scriptPath := path + "/" + file
	content, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		log.Error().Err(err).Str("script", scriptPath).Msg("read scrip failed")
		return err
	}
	scriptMap[file] = redis.NewScript(keyCount, string(content))

	return nil
}

// execute ...
func execute(file string, args ...interface{}) (reply interface{}, err error) {
	entry := time.Now()
	defer func() {
		es := time.Since(entry) / time.Millisecond
		lg := log.Debug()
		if err != nil {
			lg = log.Error().Err(err)
		}
		lg.Int32("elapsed", int32(es)).Str("script", file).
			Interface("args", args).Msg("redis do")
	}()

	script, ok := scriptMap[file]
	if !ok {
		err = errors.New("lua script not exist")
		return
	}
	con := rdsCli.getConnection()
	if con.Err() != nil {
		return nil, con.Err()
	}
	defer con.Close()
	reply, err = script.Do(con, args...)

	return reply, err
}
