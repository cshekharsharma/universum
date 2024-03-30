package engine

import (
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

func (s *Storage) Get(key string) (*entity.Record, uint32) {
	pointer, ok := keypool[key]
	if !ok {
		return nil, RESPONSECODE_KEY_NOT_FOUND
	}

	record := memoryStore[pointer]
	if record == nil {
		return nil, RESPONSECODE_KEY_NOT_FOUND
	}

	if s.hasExpired(record) {
		DeleteByPointer(record, pointer)
		return nil, RESPONSECODE_KEY_EXPIRED
	}

	record.LastAccessedAt = utils.GetCurrentEPochTime()
	return record, RESPONSECODE_RECORD_FOUND
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

	return true, RESPONSECODE_UPDATED
}

func (s *Storage) Delete() *entity.Record { return nil }

func DeleteByPointer(record *entity.Record, pointer unsafe.Pointer) bool {
	delete(memoryStore, pointer)
	delete(keypool, *((*string)(pointer)))
	delete(expirationMap, record)

	return true
}

func (s *Storage) Exists() *entity.Record { return nil }

func (s *Storage) hasExpired(record *entity.Record) bool {
	return false
}
