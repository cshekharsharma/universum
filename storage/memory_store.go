package storage

import (
	"hash/fnv"
	"sync"
	"time"
	"universum/config"
	"universum/consts"
	"universum/utils"
)

const (
	ShardCount         uint32 = 3
	infiniteExpiryTime int64  = 4102444800
)

type MemoryStore struct {
	shards [ShardCount]*Shard
}

func CreateNewMemoryStore() *MemoryStore {
	store := &MemoryStore{}
	for i := range store.shards {
		store.shards[i] = NewShard(int64(i))
	}

	var counter uint32 = 0

	for counter = 0; uint32(counter) < ShardCount; counter++ {
		store.shards[counter] = &Shard{
			id:   int64(counter),
			data: &sync.Map{},
		}
	}

	return store
}

func (ms *MemoryStore) Initialize() {}

func (ms *MemoryStore) GetAllShards() [ShardCount]*Shard {
	return ms.shards
}

func (ms *MemoryStore) Exists(key string) (bool, uint32) {
	shard := ms.getShard(key)
	val, ok := shard.data.Load(key)

	if !ok {
		return false, consts.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.IsExpired() {
		shard.data.Delete(key)
		return false, consts.CRC_RECORD_EXPIRED
	}

	return true, consts.CRC_RECORD_FOUND
}

func (ms *MemoryStore) Get(key string) (Record, uint32) {
	shard := ms.getShard(key)
	val, ok := shard.data.Load(key)

	if !ok {
		return nil, consts.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.IsExpired() {
		shard.data.Delete(key)
		return nil, consts.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, consts.CRC_RECORD_FOUND
}

func (ms *MemoryStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	record := &ScalarRecord{
		Value:  value,
		Type:   GetTypeEncoding(value),
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: infiniteExpiryTime,
	}

	if ttl > 0 {
		record.Expiry = time.Now().Unix() + ttl
	}

	shard := ms.getShard(key)
	shard.data.Store(key, record)
	return true, consts.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) Delete(key string) (bool, uint32) {
	shard := ms.getShard(key)
	shard.data.Delete(key)
	return true, consts.CRC_RECORD_DELETED
}

func (ms *MemoryStore) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := ms.Get(key)

	if code != consts.CRC_RECORD_FOUND {
		return config.INVALID_NUMERIC_VALUE, consts.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if !utils.IsInteger(record.Value) {
		return config.INVALID_NUMERIC_VALUE, consts.CRC_INCR_INVALID_TYPE
	}

	var newValue int64
	oldValue := record.Value.(int64)

	if isIncr {
		newValue = int64(oldValue) + offset
	} else {
		newValue = int64(oldValue) - offset
	}

	ttl := record.Expiry - time.Now().Unix()
	didSet, setcode := ms.Set(key, newValue, ttl)

	if !didSet {
		return config.INVALID_NUMERIC_VALUE, setcode
	}

	return newValue, consts.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) Append(key string, value string) (int64, uint32) {
	val, code := ms.Get(key)

	if code != consts.CRC_RECORD_FOUND {
		return config.INVALID_NUMERIC_VALUE, consts.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.Type != TYPE_ENCODING_STRING {
		return config.INVALID_NUMERIC_VALUE, consts.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - time.Now().Unix()

	didSet, setcode := ms.Set(key, newValue, ttl)
	if !didSet {
		return config.INVALID_NUMERIC_VALUE, setcode
	}

	return int64(len(newValue)), consts.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) getShard(key string) *Shard {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))

	shardIndex := hasher.Sum32() % ShardCount
	return ms.shards[shardIndex]
}
