package lsm

import (
	"universum/config"
	"universum/entity"
	"universum/storage/lsm/memtable"
	"universum/storage/lsm/sstable"
)

type LSMStore struct {
	memTable memtable.MemTable
	sstables []*sstable.SSTable // or another suitable data structure
	// Other fields as needed
}

func CreateNewLSMStore(mtype string) *LSMStore {
	memtable := memtable.CreateNewMemTable(config.GetMemtableStorageType())
	return &LSMStore{
		memTable: memtable,
		sstables: make([]*sstable.SSTable, 10),
	}

}
func (lsm *LSMStore) Initialize() {}

func (lsm *LSMStore) GetStoreType() string {
	return config.StorageTypeLSM
}

func (lsm *LSMStore) Exists(key string) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) Get(key string) (entity.Record, uint32) {
	// Implementation here
	return &entity.ScalarRecord{}, 0
}

func (lsm *LSMStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) Delete(key string) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) Append(key string, value string) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) MGet(keys []string) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) MDelete(keys []string) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) TTL(key string) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) Expire(key string, ttl int64) (bool, uint32) {
	// Implementation here
	return false, 0
}
