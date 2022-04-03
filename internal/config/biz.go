package config

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
)

const (
	BizConfKey = "biz"
)

type BizConf struct {
	sync.RWMutex
	ThirdParty string `json:"thirdParty"`
}

func (biz *BizConf) OnConfigChanged(key string, data []byte) error {
	biz.Lock()
	defer biz.Unlock()

	switch key {
	case BizConfKey:
		return json.Unmarshal(data, biz)
	default:
		return errors.Errorf("unknwon key '%v'", key)
	}
}

func (biz *BizConf) GetThirdParty() string {
	biz.RLock()
	defer biz.RUnlock()

	if biz.ThirdParty == "" {
		return "http://httpbin.org"
	}

	return biz.ThirdParty
}
