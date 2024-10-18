package memtable

import (
	"sync"
	"time"
	"universum/config"
	"universum/dslib"
	"universum/entity"
	"universum/internal/logger"
	"universum/utils"
)

type ListBloomMemTable struct {
	skipList    *dslib.SkipList
	bloomFilter *dslib.BloomFilter
	lock        sync.RWMutex
	size        int64
	sizeMap     sync.Map
	maxSize     int64

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
		maxSize:     config.Store.Storage.LSM.WriteBufferSize,
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

	found, _, expiry, state := m.skipList.Get(key)

	if !found {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := &entity.ScalarRecord{
		Value:  nil,
		Expiry: expiry,
		State:  state,
	}

	if found && record.IsTombstoned() {
		return false, entity.CRC_RECORD_TOMBSTONED
	}

	if found && record.IsExpired() {
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

	found, val, expiry, state := m.skipList.Get(key)
	if !found {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	record := &entity.ScalarRecord{
		Value:  val,
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: expiry,
		State:  state,
	}

	if found && record.IsTombstoned() {
		return nil, entity.CRC_RECORD_TOMBSTONED
	}

	if found && record.IsExpired() {
		m.skipList.Remove(key)
		m.reduceMemtableSize(key)
		return nil, entity.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, entity.CRC_RECORD_FOUND
}

func (m *ListBloomMemTable) Set(key string, value interface{}, ttl int64, state uint8) (bool, uint32) {
	if !utils.IsWriteableDatatype(value) {
		return false, entity.CRC_INVALID_DATATYPE
	}

	if !utils.IsWriteableDataSize(value, config.Store.Storage.MaxRecordSizeInBytes) {
		return false, entity.CRC_RECORD_TOO_BIG
	}

	expiry := config.InfiniteExpiryTime

	if ttl > 0 {
		expiry = utils.GetCurrentEPochTime() + ttl
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.IsFull() {
		m.Truncate()
	}

	m.skipList.Insert(key, value, expiry, state)
	m.updateMemtableSize(key, value)
	m.bloomFilter.Add(key)

	return true, entity.CRC_RECORD_UPDATED
}

func (m *ListBloomMemTable) Delete(key string) (bool, uint32) {
	m.Set(key, 0, 0, entity.RecordStateTombstoned)
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
	didSet, setcode := m.Set(key, newValue, ttl, entity.RecordStateActive)

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

	if !utils.IsString(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	didSet, setcode := m.Set(key, newValue, ttl, entity.RecordStateActive)
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
				"Value": record.GetValue(),
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
		didSet, _ := m.Set(key, value, 0, entity.RecordStateActive)
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
	return m.Set(key, record.Value, ttl, entity.RecordStateActive)
}

func (m *ListBloomMemTable) GetSize() int64 {
	return m.size
}

func (m *ListBloomMemTable) IsFull() bool {
	return m.size >= m.maxSize
}

func (m *ListBloomMemTable) GetCount() int64 {
	return int64(m.skipList.Size())
}

func (m *ListBloomMemTable) updateMemtableSize(key string, val interface{}) {
	var metadataSize int64 = 2 * entity.Int64SizeInBytes // 8 bytes each for exp and state
	newSize, err := utils.GetInMemorySizeInBytes(val)
	newSize += int64(len(key)) + metadataSize

	delta := newSize

	if err != nil {
		return // lets not do anything if we dont know the size
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

func (m *ListBloomMemTable) GetAll() []*entity.RecordKV {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.skipList.GetAllRecords()
}

func (m *ListBloomMemTable) Truncate() error {
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
	WALRotaterChan <- time.Now().UnixNano()

	logger.Get().Info("Memtable truncated after size=%d, count=%d",
		backupMemtable.size, backupMemtable.skipList.Size())

	return nil
}
