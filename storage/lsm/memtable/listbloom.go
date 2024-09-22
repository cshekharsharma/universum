package memtable

import (
	"sync"
	"universum/config"
	"universum/dslib"
	"universum/entity"
	"universum/utils"
)

type ListBloomMemTable struct {
	skipList    *dslib.SkipList
	bloomFilter *dslib.BloomFilter
	lock        sync.RWMutex
	size        int64
	sizeMap     sync.Map
}

func NewListBloomMemTable(maxRecords uint64, falsePositiveRate float64) *ListBloomMemTable {
	bfSize, bfHashCount := dslib.OptimalBloomFilterSize(maxRecords, falsePositiveRate)
	return &ListBloomMemTable{
		skipList:    dslib.NewSkipList(),
		bloomFilter: dslib.NewBloomFilter(bfSize, bfHashCount),
		size:        0,
	}
}

func (lb *ListBloomMemTable) Exists(key string) (bool, uint32) {
	lb.lock.RLock()
	defer lb.lock.RUnlock()

	if !lb.bloomFilter.Exists(key) {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	found, _, expiry := lb.skipList.Get(key)

	if !found {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	if found && (expiry == 0 || expiry < utils.GetCurrentEPochTime()) {
		lb.skipList.Remove(key)
		lb.reduceMemtableSize(key)
		return false, entity.CRC_RECORD_EXPIRED
	}

	return true, entity.CRC_RECORD_FOUND
}

func (lb *ListBloomMemTable) Get(key string) (entity.Record, uint32) {
	lb.lock.RLock()
	defer lb.lock.RUnlock()

	if !lb.bloomFilter.Exists(key) {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	found, val, expiry := lb.skipList.Get(key)
	if !found {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	record := &entity.ScalarRecord{
		Value:  val,
		Type:   utils.GetTypeEncoding(val),
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: expiry,
	}

	if found && record.IsExpired() {
		lb.skipList.Remove(key)
		lb.reduceMemtableSize(key)
		return nil, entity.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, entity.CRC_RECORD_FOUND
}

func (lb *ListBloomMemTable) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	if !utils.IsWriteableDatatype(value) {
		return false, entity.CRC_INVALID_DATATYPE
	}

	if !utils.IsWriteableDataSize(value, config.GetMaxRecordSizeInBytes()) {
		return false, entity.CRC_RECORD_TOO_BIG
	}

	expiry := infiniteExpiryTime

	if ttl > 0 {
		expiry = utils.GetCurrentEPochTime() + ttl
	}

	lb.lock.Lock()
	defer lb.lock.Unlock()

	lb.skipList.Insert(key, value, expiry)
	lb.updateMemtableSize(key, value)
	lb.bloomFilter.Add(key)

	return true, entity.CRC_RECORD_UPDATED
}

func (lb *ListBloomMemTable) Delete(key string) (bool, uint32) {
	lb.lock.Lock()
	defer lb.lock.Unlock()

	lb.skipList.Remove(key)
	lb.reduceMemtableSize(key)

	return true, entity.CRC_RECORD_DELETED
}

func (lb *ListBloomMemTable) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := lb.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
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
	didSet, setcode := lb.Set(key, newValue, ttl)

	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return newValue, entity.CRC_RECORD_UPDATED
}

func (lb *ListBloomMemTable) Append(key string, value string) (int64, uint32) {
	val, code := lb.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
	if record.Type != utils.TYPE_ENCODING_STRING {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	didSet, setcode := lb.Set(key, newValue, ttl)
	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return int64(len(newValue)), entity.CRC_RECORD_UPDATED
}

func (lb *ListBloomMemTable) MGet(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		record, code := lb.Get(keys[idx])

		if _, ok := record.(*entity.ScalarRecord); ok {
			responseMap[keys[idx]] = map[string]interface{}{
				"Value": record.(*entity.ScalarRecord).Value,
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

func (lb *ListBloomMemTable) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for key, value := range kvMap {
		didSet, _ := lb.Set(key, value, 0)
		responseMap[key] = didSet
	}

	return responseMap, entity.CRC_MSET_COMPLETED
}

func (lb *ListBloomMemTable) MDelete(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		deleted, _ := lb.Delete(keys[idx])
		responseMap[keys[idx]] = deleted
	}

	return responseMap, entity.CRC_MDEL_COMPLETED
}

func (lb *ListBloomMemTable) TTL(key string) (int64, uint32) {
	val, code := lb.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return 0, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	return ttl, entity.CRC_RECORD_FOUND
}

func (lb *ListBloomMemTable) Expire(key string, ttl int64) (bool, uint32) {
	val, code := lb.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
	return lb.Set(key, record.Value, ttl)
}

func (lb *ListBloomMemTable) GetSize() int64 {
	return lb.size
}

func (lb *ListBloomMemTable) IsFull() bool {
	return lb.size >= DefaultMemTableSize
}

func (lb *ListBloomMemTable) GetRecordCount() int64 {
	return int64(lb.skipList.Size())
}

func (lb *ListBloomMemTable) updateMemtableSize(key string, val interface{}) {
	var expiryValSize int64 = 1 << 3 // 8 bytes for expiry info
	newSize, err := utils.GetSizeInBytes(val)
	newSize += int64(len(key)) + expiryValSize

	delta := newSize

	if err != nil {
		return // lets not anything if we dont know the size
	}

	prevSize, ok := lb.sizeMap.Load(key)
	if ok {
		// assuming previous size also includes size of key and expiry-info
		delta = int64(newSize) - prevSize.(int64)
	}

	lb.sizeMap.Store(key, newSize)
	lb.size += delta
}

func (lb *ListBloomMemTable) reduceMemtableSize(key string) {
	prevSize, ok := lb.sizeMap.Load(key)
	if ok {
		lb.sizeMap.Delete(key)
		lb.size -= prevSize.(int64)
	}

	if lb.size < 0 {
		lb.size = 0
	}
}
