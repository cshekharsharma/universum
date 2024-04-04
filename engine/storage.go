package engine

import (
	"sync"
	"universum/config"
	"universum/consts"
	"universum/engine/entity"
	"universum/utils"
	"unsafe"
)

var memoryStore map[unsafe.Pointer]*entity.Record
var memStoreMutex sync.RWMutex

var expirationMap map[*entity.Record]uint32
var expirationMutex sync.RWMutex

var keypool map[string]unsafe.Pointer
var keypoolMutex sync.RWMutex

type Storage struct {
}

func (s *Storage) Initialize() {
	memoryStore = make(map[unsafe.Pointer]*entity.Record)
	expirationMap = make(map[*entity.Record]uint32)
	keypool = make(map[string]unsafe.Pointer)
}

func (s *Storage) Exists(key string) (bool, uint32) {
	keypoolMutex.RLock()
	pointer, ok := keypool[key]
	keypoolMutex.RUnlock()

	if !ok {
		return false, consts.CRC_RECORD_NOT_FOUND
	}

	memStoreMutex.RLock()
	record, ok := memoryStore[pointer]
	memStoreMutex.RUnlock()

	if ok {
		expired := isExpiredRecord(record)

		if !expired {
			return true, consts.CRC_RECORD_FOUND
		} else {
			deleteByPointer(record, pointer)
			return false, consts.CRC_RECORD_EXPIRED
		}
	}

	return false, consts.CRC_RECORD_NOT_FOUND
}

func (s *Storage) Get(key string) (*entity.Record, uint32) {
	keypoolMutex.RLock()
	pointer, ok := keypool[key]
	keypoolMutex.RUnlock()

	if !ok {
		return nil, consts.CRC_RECORD_NOT_FOUND
	}

	memStoreMutex.RLock()
	record := memoryStore[pointer]
	memStoreMutex.RUnlock()

	if record == nil {
		return nil, consts.CRC_RECORD_NOT_FOUND
	}

	if isExpiredRecord(record) {
		deleteByPointer(record, pointer)
		return nil, consts.CRC_RECORD_EXPIRED
	}

	record.LAT = utils.GetCurrentEPochTime()
	return record, consts.CRC_RECORD_FOUND
}

func (s *Storage) Set(key string, value interface{}, ttl uint32) (bool, uint32) {
	//TriggerPeriodicExpiredRecordCleaup()
	record := entity.NewRecord(value, GetTypeEncoding(value), utils.GetCurrentEPochTime())

	keypoolMutex.RLock()
	pointer, ok := keypool[key]
	keypoolMutex.RUnlock()

	if !ok {
		pointer = unsafe.Pointer(&key)

		keypoolMutex.Lock()
		keypool[key] = pointer
		keypoolMutex.Unlock()
	}

	memStoreMutex.Lock()
	memoryStore[pointer] = record
	memStoreMutex.Unlock()

	if ttl > 0 {
		expirationMutex.Lock()
		expirationMap[record] = utils.GetCurrentEPochTime() + ttl
		expirationMutex.Unlock()
	}

	return true, consts.CRC_RECORD_UPDATED
}

func (s *Storage) Delete(key string) (bool, uint32) {
	if exists, _ := s.Exists(key); !exists {
		return true, consts.CRC_RECORD_DELETED
	}

	keypoolMutex.RLock()
	pointer, ok := keypool[key]
	keypoolMutex.RUnlock()

	if ok {
		memStoreMutex.RLock()
		record := memoryStore[pointer]
		memStoreMutex.RUnlock()

		deleted := deleteByPointer(record, pointer)

		if deleted {
			return true, consts.CRC_RECORD_DELETED
		}
	}

	return false, consts.CRC_RECORD_NOT_DELETED
}

func (s *Storage) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	record, code := s.Get(key)

	if code == consts.CRC_RECORD_FOUND {
		if utils.IsInteger(record.Value) {
			var newValue int64
			oldValue := record.Value.(int64)

			if isIncr {
				newValue = int64(oldValue) + offset
			} else {
				newValue = int64(oldValue) - offset
			}

			var ttl uint32 = 0

			expirationMutex.RLock()
			expiry, ok := expirationMap[record]
			expirationMutex.RUnlock()

			if ok {
				ttl = expiry - utils.GetCurrentEPochTime()
			}

			didSet, setcode := s.Set(key, newValue, ttl)
			if didSet {
				return newValue, consts.CRC_RECORD_UPDATED
			}

			return config.INVALID_NUMERIC_VALUE, setcode
		}

		return config.INVALID_NUMERIC_VALUE, consts.CRC_INCR_INVALID_TYPE
	}

	return config.INVALID_NUMERIC_VALUE, consts.CRC_RECORD_NOT_FOUND
}

func (s *Storage) Append(key string, offset int64) (*entity.Record, uint32) {
	return nil, 1
}

// ------------------- Internal Functions -------------------

func deleteByPointer(record *entity.Record, pointer unsafe.Pointer) bool {
	memStoreMutex.Lock()
	delete(memoryStore, pointer)
	memStoreMutex.Unlock()

	keypoolMutex.Lock()
	delete(keypool, *((*string)(pointer)))
	keypoolMutex.Unlock()

	expirationMutex.Lock()
	delete(expirationMap, record)
	expirationMutex.Unlock()

	return true
}
