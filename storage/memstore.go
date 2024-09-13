package storage

import (
	"hash/fnv"
	"sync"
	"universum/config"
	"universum/entity"
	"universum/utils"
)

const (
	ShardCount         uint32 = 32
	infiniteExpiryTime int64  = 4102444800
)

type MemoryStore struct {
	shards [ShardCount]*Shard
}

func createNewMemoryStore() *MemoryStore {
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
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.IsExpired() {
		shard.data.Delete(key)
		return false, entity.CRC_RECORD_EXPIRED
	}

	return true, entity.CRC_RECORD_FOUND
}

func (ms *MemoryStore) Get(key string) (Record, uint32) {
	shard := ms.getShard(key)
	val, ok := shard.data.Load(key)

	if !ok {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.IsExpired() {
		shard.data.Delete(key)
		return nil, entity.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, entity.CRC_RECORD_FOUND
}

func (ms *MemoryStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	record := &ScalarRecord{
		Value:  value,
		Type:   utils.GetTypeEncoding(value),
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: infiniteExpiryTime,
	}

	if !utils.IsWriteableDatatype(value) {
		return false, entity.CRC_INVALID_DATATYPE
	}

	if !utils.IsWriteableDataSize(value, config.GetMaxRecordSizeInBytes()) {
		return false, entity.CRC_RECORD_TOO_BIG
	}

	if ttl > 0 {
		record.Expiry = utils.GetCurrentEPochTime() + ttl
	}

	shard := ms.getShard(key)
	shard.data.Store(key, record)
	return true, entity.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) Delete(key string) (bool, uint32) {
	shard := ms.getShard(key)
	shard.data.Delete(key)
	return true, entity.CRC_RECORD_DELETED
}

func (ms *MemoryStore) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := ms.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if !utils.IsInteger(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	var newValue int64
	oldValue := record.Value.(int64)

	if isIncr {
		newValue = int64(oldValue) + offset
	} else {
		newValue = int64(oldValue) - offset
	}

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	didSet, setcode := ms.Set(key, newValue, ttl)

	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return newValue, entity.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) Append(key string, value string) (int64, uint32) {
	val, code := ms.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	if record.Type != utils.TYPE_ENCODING_STRING {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	didSet, setcode := ms.Set(key, newValue, ttl)
	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return int64(len(newValue)), entity.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) MGet(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		record, code := ms.Get(keys[idx])

		if _, ok := record.(*ScalarRecord); ok {
			responseMap[keys[idx]] = map[string]interface{}{
				"Value": record.(*ScalarRecord).Value,
				"Code":  code,
			}
		} else {
			responseMap[keys[idx]] = map[string]interface{}{
				"Value": nil,
				"Code":  code,
			}
		}
	}

	return responseMap, entity.CRC_MGET_COMPLETED
}

func (ms *MemoryStore) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for key, value := range kvMap {
		didSet, _ := ms.Set(key, value, 0)
		responseMap[key] = didSet
	}

	return responseMap, entity.CRC_MSET_COMPLETED
}

func (ms *MemoryStore) MDelete(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		deleted, _ := ms.Delete(keys[idx])
		responseMap[keys[idx]] = deleted
	}

	return responseMap, entity.CRC_MDEL_COMPLETED
}

func (ms *MemoryStore) TTL(key string) (int64, uint32) {
	val, code := ms.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return 0, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	if ttl == infiniteExpiryTime {
		ttl = -1
	}

	return ttl, entity.CRC_RECORD_FOUND
}

func (ms *MemoryStore) Expire(key string, ttl int64) (bool, uint32) {
	val, code := ms.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*ScalarRecord)
	record.LAT = utils.GetCurrentEPochTime()
	record.Expiry = utils.GetCurrentEPochTime() + ttl

	if ttl == 0 {
		record.Expiry = infiniteExpiryTime
	}

	shard := ms.getShard(key)
	shard.data.Store(key, record)
	return true, entity.CRC_RECORD_UPDATED
}

func (ms *MemoryStore) getShard(key string) *Shard {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))

	shardIndex := hasher.Sum32() % ShardCount
	return ms.shards[shardIndex]
}
