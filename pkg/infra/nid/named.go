package nid

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/pkg/errors"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
	nodePrefix = "node_"
	retryCount = 5
	bucketName = "nodeId"
)

func init() {
	consul.Register()
}

type NodeNamed interface {
	GetNodeID(*NameHolder) (int, error)
}

// NameHolder ...
type NameHolder struct {
	LocalPath  string `json:"localPath"`
	LocalIP    string `json:"localIp"`
	ApplyTime  string `json:"applyTime"`
	ServiceKey string `json:"-"`
}

func (h *NameHolder) decode(data []byte) error {
	err := json.Unmarshal(data, h)
	return err
}

func (h *NameHolder) encode() ([]byte, error) {
	return json.Marshal(h)
}

func NewConsulNamed(addr string) (NodeNamed, error) {
	kvStore, err := libkv.NewStore(
		store.CONSUL,
		[]string{addr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)

	if err != nil {
		return nil, err
	}

	return &nodeNamed{
		Store:      kvStore,
		retryCount: retryCount,
	}, nil
}

func NewEtcdNamed(addr string) (NodeNamed, error) {
	kvStore, err := libkv.NewStore(
		store.ETCD,
		[]string{addr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)

	if err != nil {
		return nil, err
	}

	return &nodeNamed{
		Store:      kvStore,
		retryCount: retryCount,
	}, nil
}

func NewBoltNamed(addr string) (NodeNamed, error) {
	kvStore, err := libkv.NewStore(
		store.BOLTDB,
		[]string{addr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
			Bucket:            bucketName,
		},
	)

	if err != nil {
		return nil, err
	}

	return &nodeNamed{
		Store:      kvStore,
		retryCount: retryCount,
	}, nil
}

type nodeNamed struct {
	store.Store
	retryCount int
}

func (c *nodeNamed) GetNodeID(holder *NameHolder) (nodeID int, err error) {
	holder.LocalPath, _ = filepath.Abs(holder.LocalPath)
	nodeID, err = c.recoverNodeID(holder)
	if err != nil {
		return
	}

	if nodeID == 0 {
		nodeID, err = c.applyNodeID(holder)
	}

	return
}

// RecoverNodeID 恢复配置
func (c *nodeNamed) recoverNodeID(holder *NameHolder) (int, error) {
	kvPairs, err := c.List(holder.ServiceKey)
	if err != nil {
		if err != store.ErrKeyNotFound {
			return 0, err
		}
	}

	for _, pair := range kvPairs {
		info := &NameHolder{}
		if info.decode(pair.Value) != nil ||
			info.LocalIP != holder.LocalIP ||
			info.LocalPath != holder.LocalPath {
			continue
		}

		if err := c.tryHold(pair, holder); err != nil {
			return 0, err
		}

		return c.convertStringToID(pair.Key), nil
	}
	return 0, nil
}

// ApplyNodeID 申请配置
func (c *nodeNamed) applyNodeID(holder *NameHolder) (int, error) {
	for i := 0; i < c.retryCount; i++ {
		pairs, err := c.List(holder.ServiceKey)
		if err != nil {
			if err != store.ErrKeyNotFound {
				return 0, err
			}
		}

		newID := c.makeNewID(pairs)
		if err := c.tryHold(&store.KVPair{
			Key:       c.makeConsulKey(holder.ServiceKey, newID),
			LastIndex: 0,
		}, holder); err == nil {
			return newID, nil
		}
	}
	return 0, errors.Errorf("try to hold %d times, but failed", c.retryCount)
}

func (c *nodeNamed) makeNewID(pairs []*store.KVPair) int {
	usedIDs := make([]int, 256)
	for _, pair := range pairs {
		nid := c.convertStringToID(pair.Key)
		if nid >= len(usedIDs) {
			tmp := usedIDs
			usedIDs := make([]int, nid*2)
			copy(usedIDs, tmp)
		}
		usedIDs[nid] = nid
	}

	newID := 1
	for ; newID < len(usedIDs); newID++ {
		if usedIDs[newID] == 0 {
			break
		}
	}

	return newID
}

func (c *nodeNamed) convertStringToID(s string) int {
	paths := strings.Split(s, "/")
	length := len(paths)
	if length == 0 {
		return 0
	}

	id, err := strconv.Atoi(strings.TrimPrefix(paths[length-1], nodePrefix))
	if err != nil || id < 0 {
		return 0
	}
	return id
}

func (c *nodeNamed) makeConsulKey(nodeKeyPrefix string, id int) string {
	return fmt.Sprintf("%v/nodeId/%v%02v", nodeKeyPrefix, nodePrefix, id)
}

func (c *nodeNamed) tryHold(pair *store.KVPair, holder *NameHolder) error {
	newPair, err := c.Get(pair.Key)
	if err != nil {
		if err != store.ErrKeyNotFound {
			return err
		}
	} else {
		if newPair.LastIndex > pair.LastIndex {
			return errors.New("try hold failed")
		}
	}

	holder.ApplyTime = time.Now().Format(timeFormat)
	pair.Value, err = holder.encode()
	if err != nil {
		return err
	}

	if newPair == nil {
		_, _, err = c.AtomicPut(pair.Key, pair.Value, nil, nil)
	} else {
		_, _, err = c.AtomicPut(pair.Key, pair.Value, pair, nil)
	}

	return err
}
