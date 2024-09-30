package memtable

import (
	"universum/config"
	"universum/entity"
)

const (
	DefaultMemTableSize int64 = 64 * 1024 * 1024 // 64 MB
	infiniteExpiryTime  int64 = 4102444800
)

type MemTable interface {
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
	GetSize() int64
	IsFull() bool
	GetRecordCount() int64
	GetAllRecords() map[string]interface{}
}

func CreateNewMemTable(tabletype string) MemTable {
	switch tabletype {
	case config.MemtableStorageTypeLB: // implementated with `skiplist + bloom filter`
		return &ListBloomMemTable{}

	case config.MemtableStorageTypeLM: // implemented with `map + sorted list`
		return &ListMapMemTable{}

	default:
		return &ListBloomMemTable{}
	}
}