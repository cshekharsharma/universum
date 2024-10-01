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

	// bloom filter size and hash count
	bfSize      uint64
	bfHashCount uint8
}

func NewListBloomMemTable(maxRecords int64, falsePositiveRate float64) *ListBloomMemTable {
	bfSize, bfHashCount := dslib.OptimalBloomFilterSize(maxRecords, falsePositiveRate)
	return &ListBloomMemTable{
		skipList:    dslib.NewSkipList(),
		bloomFilter: dslib.NewBloomFilter(bfSize, bfHashCount),
		size:        0,
		bfSize:      bfSize,
		bfHashCount: bfHashCount,
	}
}

func (m *ListBloomMemTable) Exists(key string) (bool, uint32) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if !m.bloomFilter.Exists(key) {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	found, _, expiry := m.skipList.Get(key)

	if !found {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	if found && (expiry == 0 || expiry < utils.GetCurrentEPochTime()) {
		m.skipList.Remove(key)
		m.reduceMemtableSize(key)
		return false, entity.CRC_RECORD_EXPIRED
	}

	return true, entity.CRC_RECORD_FOUND
}

func (m *ListBloomMemTable) Get(key string) (entity.Record, uint32) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if !m.bloomFilter.Exists(key) {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	found, val, expiry := m.skipList.Get(key)
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
		m.skipList.Remove(key)
		m.reduceMemtableSize(key)
		return nil, entity.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, entity.CRC_RECORD_FOUND
}

func (m *ListBloomMemTable) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	if !utils.IsWriteableDatatype(value) {
		return false, entity.CRC_INVALID_DATATYPE
	}

	if !utils.IsWriteableDataSize(value, config.Store.Storage.MaxRecordSizeInBytes) {
		return false, entity.CRC_RECORD_TOO_BIG
	}

	expiry := infiniteExpiryTime

	if ttl > 0 {
		expiry = utils.GetCurrentEPochTime() + ttl
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.IsFull() {
		m.TruncateMemtable()
	}

	m.skipList.Insert(key, value, expiry)
	m.updateMemtableSize(key, value)
	m.bloomFilter.Add(key)

	return true, entity.CRC_RECORD_UPDATED
}

func (m *ListBloomMemTable) Delete(key string) (bool, uint32) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.skipList.Remove(key)
	m.reduceMemtableSize(key)

	return true, entity.CRC_RECORD_DELETED
}

func (m *ListBloomMemTable) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := m.Get(key)

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
	didSet, setcode := m.Set(key, newValue, ttl)

	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return newValue, entity.CRC_RECORD_UPDATED
}

func (m *ListBloomMemTable) Append(key string, value string) (int64, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
	if record.Type != utils.TYPE_ENCODING_STRING {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	didSet, setcode := m.Set(key, newValue, ttl)
	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return int64(len(newValue)), entity.CRC_RECORD_UPDATED
}

func (m *ListBloomMemTable) MGet(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		record, code := m.Get(keys[idx])

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

func (m *ListBloomMemTable) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for key, value := range kvMap {
		didSet, _ := m.Set(key, value, 0)
		responseMap[key] = didSet
	}

	return responseMap, entity.CRC_MSET_COMPLETED
}

func (m *ListBloomMemTable) MDelete(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		deleted, _ := m.Delete(keys[idx])
		responseMap[keys[idx]] = deleted
	}

	return responseMap, entity.CRC_MDEL_COMPLETED
}

func (m *ListBloomMemTable) TTL(key string) (int64, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return 0, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	return ttl, entity.CRC_RECORD_FOUND
}

func (m *ListBloomMemTable) Expire(key string, ttl int64) (bool, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
	return m.Set(key, record.Value, ttl)
}

func (m *ListBloomMemTable) GetSize() int64 {
	return m.size
}

func (m *ListBloomMemTable) IsFull() bool {
	return m.size >= DefaultMemTableSize
}

func (m *ListBloomMemTable) GetRecordCount() int64 {
	return int64(m.skipList.Size())
}

func (m *ListBloomMemTable) updateMemtableSize(key string, val interface{}) {
	var expiryValSize int64 = 1 << 3 // 8 bytes for expiry info
	newSize, err := utils.GetInMemorySizeInBytes(val)
	newSize += int64(len(key)) + expiryValSize

	delta := newSize

	if err != nil {
		return // lets not anything if we dont know the size
	}

	prevSize, ok := m.sizeMap.Load(key)
	if ok {
		// assuming previous size also includes size of key and expiry-info
		delta = int64(newSize) - prevSize.(int64)
	}

	m.sizeMap.Store(key, newSize)
	m.size += delta
}

func (m *ListBloomMemTable) reduceMemtableSize(key string) {
	prevSize, ok := m.sizeMap.Load(key)
	if ok {
		m.sizeMap.Delete(key)
		m.size -= prevSize.(int64)
	}

	if m.size < 0 {
		m.size = 0
	}
}

func (m *ListBloomMemTable) GetAllRecords() map[string]entity.Record {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.skipList.GetAllRecords()
}

func (m *ListBloomMemTable) TruncateMemtable() {
	backupMemtable := &ListBloomMemTable{
		skipList:    m.skipList,
		bloomFilter: m.bloomFilter,
		size:        m.size,
		sizeMap:     sync.Map{},
	}

	m.skipList = dslib.NewSkipList()
	m.bloomFilter = dslib.NewBloomFilter(m.bloomFilter.Size, m.bloomFilter.HashCount)
	m.size = 0
	m.sizeMap = sync.Map{}

	FlusherChan <- backupMemtable
}
