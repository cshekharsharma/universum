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

type TreeBloomMemTable struct {
	rbTree      *dslib.RBTree
	bloomFilter *dslib.BloomFilter
	lock        sync.RWMutex
	size        int64
	sizeMap     sync.Map
	maxSize     int64

	// Bloom Filter configuration
	bfSize      uint64
	bfHashCount uint8
}

// NewTreeBloomMemTable initializes a new TreeBloomMemTable with a specified maximum record count and false positive rate.
func NewTreeBloomMemTable(maxRecords int64, falsePositiveRate float64) *TreeBloomMemTable {
	bfSize, bfHashCount := dslib.OptimalBloomFilterSize(maxRecords, falsePositiveRate)
	return &TreeBloomMemTable{
		rbTree:      dslib.NewRBTree(),
		bloomFilter: dslib.NewBloomFilter(bfSize, bfHashCount),
		maxSize:     config.Store.Storage.LSM.WriteBufferSize,
		bfSize:      bfSize,
		bfHashCount: bfHashCount,
	}
}

// Exists checks if a key exists in the memtable using the Bloom Filter.
func (m *TreeBloomMemTable) Exists(key string) (bool, uint32) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if !m.bloomFilter.Exists(key) {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	found, val, expiry, state := m.rbTree.Get(key)
	if !found {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := &entity.ScalarRecord{
		Value:  val,
		Expiry: expiry,
		State:  state,
	}

	if record.IsTombstoned() {
		return false, entity.CRC_RECORD_TOMBSTONED
	}

	if record.IsExpired() {
		m.rbTree.Delete(key)
		m.reduceMemtableSize(key)
		return false, entity.CRC_RECORD_EXPIRED
	}

	return true, entity.CRC_RECORD_FOUND
}

// Get retrieves the value associated with a key in the memtable.
func (m *TreeBloomMemTable) Get(key string) (entity.Record, uint32) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if !m.bloomFilter.Exists(key) {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	found, val, expiry, state := m.rbTree.Get(key)
	if !found {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	record := &entity.ScalarRecord{
		Value:  val,
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: expiry,
		State:  state,
	}

	if record.IsTombstoned() {
		return nil, entity.CRC_RECORD_TOMBSTONED
	}

	if record.IsExpired() {
		m.rbTree.Delete(key)
		m.reduceMemtableSize(key)
		return nil, entity.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, entity.CRC_RECORD_FOUND
}

// Set inserts or updates a key-value pair in the memtable with an optional TTL.
func (m *TreeBloomMemTable) Set(key string, value interface{}, ttl int64, state uint8) (bool, uint32) {
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

	m.rbTree.Insert(key, value, expiry, state)
	m.updateMemtableSize(key, value)
	m.bloomFilter.Add(key)

	return true, entity.CRC_RECORD_UPDATED
}

// Delete marks a key as tombstoned, effectively removing it from the memtable.
func (m *TreeBloomMemTable) Delete(key string) (bool, uint32) {
	m.Set(key, nil, 0, entity.RecordStateTombstoned)
	return true, entity.CRC_RECORD_DELETED
}

// IncrDecrInteger increments or decrements an integer key by the specified offset.
func (m *TreeBloomMemTable) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, code
	}

	record := val.(*entity.ScalarRecord)
	if !utils.IsInteger(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	var newValue int64
	oldValue := record.Value.(int64)
	if isIncr {
		newValue = oldValue + offset
	} else {
		newValue = oldValue - offset
	}

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	success, setCode := m.Set(key, newValue, ttl, entity.RecordStateActive)
	if !success {
		return config.InvalidNumericValue, setCode
	}

	return newValue, entity.CRC_RECORD_UPDATED
}

// Append appends a string value to an existing string key.
func (m *TreeBloomMemTable) Append(key string, value string) (int64, uint32) {
	val, code := m.Get(key)
	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, code
	}

	record := val.(*entity.ScalarRecord)
	if !utils.IsString(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	success, setCode := m.Set(key, newValue, ttl, entity.RecordStateActive)
	if !success {
		return config.InvalidNumericValue, setCode
	}

	return int64(len(newValue)), entity.CRC_RECORD_UPDATED
}

// MGet retrieves values for multiple keys in one call.
func (m *TreeBloomMemTable) MGet(keys []string) (map[string]interface{}, uint32) {
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

// MSet sets multiple key-value pairs in one call.
func (m *TreeBloomMemTable) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for key, value := range kvMap {
		success, _ := m.Set(key, value, 0, entity.RecordStateActive)
		responseMap[key] = success
	}
	return responseMap, entity.CRC_MSET_COMPLETED
}

// MDelete deletes multiple keys in one call.
func (m *TreeBloomMemTable) MDelete(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for _, key := range keys {
		deleted, _ := m.Delete(key)
		responseMap[key] = deleted
	}

	return responseMap, entity.CRC_MDEL_COMPLETED
}

// TTL retrieves the remaining time-to-live (TTL) of a key.
func (m *TreeBloomMemTable) TTL(key string) (int64, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return 0, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	return ttl, entity.CRC_RECORD_FOUND
}

// Expire sets a TTL on a key, after which it will be deleted.
func (m *TreeBloomMemTable) Expire(key string, ttl int64) (bool, uint32) {
	val, code := m.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	record := val.(*entity.ScalarRecord)
	return m.Set(key, record.Value, ttl, entity.RecordStateActive)
}

// GetSize returns the current size of the memtable.
func (m *TreeBloomMemTable) GetSize() int64 {
	return m.size
}

// IsFull checks if the memtable has reached its maximum size.
func (m *TreeBloomMemTable) IsFull() bool {
	return m.size >= m.maxSize
}

// GetCount returns the total number of entries in the memtable.
func (m *TreeBloomMemTable) GetCount() int64 {
	return m.rbTree.GetSize()
}

// GetAll retrieves all records from the memtable.
func (m *TreeBloomMemTable) GetAll() []*entity.RecordKV {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.rbTree.GetAllRecords()
}

// Truncate clears the memtable, freeing memory space.
func (m *TreeBloomMemTable) Truncate() error {
	backupMemtable := &TreeBloomMemTable{
		rbTree:      m.rbTree,
		bloomFilter: m.bloomFilter,
		size:        m.size,
		sizeMap:     sync.Map{},
	}

	m.rbTree = dslib.NewRBTree()
	m.bloomFilter = dslib.NewBloomFilter(m.bfSize, m.bfHashCount)
	m.size = 0
	m.sizeMap = sync.Map{}

	FlusherChan <- backupMemtable
	WALRotaterChan <- time.Now().UnixNano()

	logger.Get().Info("Memtable truncated after size=%d, count=%d",
		backupMemtable.size, backupMemtable.rbTree.GetSize())

	return nil
}

// function for size management
func (m *TreeBloomMemTable) updateMemtableSize(key string, val interface{}) {
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

// function for size management
func (m *TreeBloomMemTable) reduceMemtableSize(key string) {
	prevSize, ok := m.sizeMap.Load(key)

	if ok {
		m.sizeMap.Delete(key)
		m.size -= prevSize.(int64)
	}

	if m.size < 0 {
		m.size = 0
	}
}
