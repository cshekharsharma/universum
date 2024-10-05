package memtable

import "universum/entity"

type ListMapMemTable struct {
}

func (lm *ListMapMemTable) Exists(key string) (bool, uint32) {
	return false, 0
}
func (lm *ListMapMemTable) Get(key string) (entity.Record, uint32) {
	return &entity.ScalarRecord{}, 0
}
func (lm *ListMapMemTable) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	return false, 0
}
func (lm *ListMapMemTable) Delete(key string) (bool, uint32) {
	return false, 0
}
func (lm *ListMapMemTable) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	return 0, 0
}
func (lm *ListMapMemTable) Append(key string, value string) (int64, uint32) {
	return 0, 0
}
func (lm *ListMapMemTable) MGet(keys []string) (map[string]interface{}, uint32) {
	return make(map[string]interface{}), 0
}
func (lm *ListMapMemTable) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	return make(map[string]interface{}), 0
}
func (lm *ListMapMemTable) MDelete(keys []string) (map[string]interface{}, uint32) {
	return make(map[string]interface{}), 0
}
func (lm *ListMapMemTable) TTL(key string) (int64, uint32) {
	return 0, 0
}
func (lm *ListMapMemTable) Expire(key string, ttl int64) (bool, uint32) {
	return false, 0
}
func (lm *ListMapMemTable) GetSize() int64 {
	return 0
}
func (lm *ListMapMemTable) IsFull() bool {
	return false
}

func (lm *ListMapMemTable) GetCount() int64 {
	return 0
}

func (lb *ListMapMemTable) GetAll() []*entity.RecordKV {
	return nil
}
