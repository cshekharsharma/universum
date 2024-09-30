package storage

import (
	"universum/entity"
)

type DataStore interface {
	Initialize() error
	GetStoreType() string

	Exists(key string) (bool, uint32)
	Get(key string) (entity.Record, uint32)
	Set(key string, value interface{}, ttl int64) (bool, uint32)
	Delete(key string) (bool, uint32)
	IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32)
	Append(key string, value string) (int64, uint32)
	MGet(keys []string) (map[string]interface{}, uint32)
	MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32)
	MDelete(keys []string) (map[string]interface{}, uint32)
	TTL(key string) (int64, uint32)
	Expire(key string, ttl int64) (bool, uint32)
}

type SnapshotService interface {
	ShouldRestore() (bool, error)
	Snapshot(store DataStore) (int64, int64, error)
	Restore(store DataStore) (int64, error)
}
