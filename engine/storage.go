package engine

import (
	"universum/consts"
	"universum/engine/entity"
	"universum/utils"
	"unsafe"
)

var memoryStore map[unsafe.Pointer]*entity.Record
var expirationMap map[*entity.Record]uint64
var keypool map[string]unsafe.Pointer

type Storage struct {
}

func (s *Storage) Initialize() {
	memoryStore = make(map[unsafe.Pointer]*entity.Record)
	expirationMap = make(map[*entity.Record]uint64)
	keypool = make(map[string]unsafe.Pointer)
}

func (s *Storage) Exists(key string) (bool, uint32) {
	pointer, ok := keypool[key]
	if !ok {
		return false, consts.CRC_RECORD_NOT_FOUND
	}

	record, ok := memoryStore[pointer]
	if ok {
		expired := s.hasExpired(record)

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
	pointer, ok := keypool[key]
	if !ok {
		return nil, consts.CRC_RECORD_NOT_FOUND
	}

	record := memoryStore[pointer]
	if record == nil {
		return nil, consts.CRC_RECORD_NOT_FOUND
	}

	if s.hasExpired(record) {
		deleteByPointer(record, pointer)
		return nil, consts.CRC_RECORD_EXPIRED
	}

	record.LastAccessedAt = utils.GetCurrentEPochTime()
	return record, consts.CRC_RECORD_FOUND
}

func (s *Storage) Set(key string, value interface{}, ttl uint64) (bool, uint32) {
	// @TODO put some eviction logic here

	record := &entity.Record{
		TypeEncoding:   0,
		Value:          value,
		LastAccessedAt: utils.GetCurrentEPochTime(),
	}

	pointer, ok := keypool[key]
	if !ok {
		keypool[key] = unsafe.Pointer(&key)
		pointer = unsafe.Pointer(&key)
	}

	memoryStore[pointer] = record

	if ttl > 0 {
		expirationMap[record] = uint64(utils.GetCurrentEPochTime()) + ttl
	}

	return true, consts.CRC_UPDATED
}

func (s *Storage) Delete(key string) (bool, uint32) {
	if exists, _ := s.Exists(key); !exists {
		return true, consts.CRC_RECORD_DELETED
	}

	pointer := keypool[key]
	record := memoryStore[pointer]

	deleted := deleteByPointer(record, pointer)

	if deleted {
		return true, consts.CRC_RECORD_DELETED
	}

	return false, consts.CRC_RECORD_NOT_DELETED
}

// /////////////////////// Internal Functions ///////////////////////////
func (s *Storage) hasExpired(record *entity.Record) bool {
	expiry, ok := expirationMap[record]
	if !ok {
		return false
	}

	return expiry < uint64(utils.GetCurrentEPochTime())
}

func deleteByPointer(record *entity.Record, pointer unsafe.Pointer) bool {
	delete(memoryStore, pointer)
	delete(keypool, *((*string)(pointer)))
	delete(expirationMap, record)

	return true
}
