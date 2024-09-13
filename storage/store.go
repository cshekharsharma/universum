package storage

import (
	"fmt"
	"strings"
	"universum/config"
)

type DataStore interface {
	Initialize()
	Exists(key string) (bool, uint32)
	Get(key string) (Record, uint32)
	Set(key string, value interface{}, ttl int64) (bool, uint32)
	Delete(key string) (bool, uint32)
	IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32)
	Append(key string, value string) (int64, uint32)
	MGet(keys []string) (map[string]interface{}, uint32)
	MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32)
	MDelete(keys []string) (map[string]interface{}, uint32)
	TTL(key string) (int64, uint32)
	Expire(key string, ttl int64) (bool, uint32)
	GetAllShards() [ShardCount]*Shard
}

var allStores = make(map[string]DataStore)

func GetStore(id string) (DataStore, error) {
	id = strings.ToUpper(id) // to avoid casing typos

	switch id {
	case config.StorageTypeMemory:
		if _, ok := allStores[id]; !ok {
			allStores[config.StorageTypeMemory] = createNewMemoryStore()
		}
		return allStores[config.StorageTypeMemory], nil

	default:
		return nil, fmt.Errorf("invalid storage type %s", id)
	}
}
