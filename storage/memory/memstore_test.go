package memory

import (
	"testing"
	"time"
	"universum/config"
	"universum/entity"
)

func SetUpMemstoreTests() {
	config.Store = config.GetSkeleton()
	config.Store.Storage.StorageEngine = config.StorageEngineMemory
	config.Store.Storage.Memory.AllowedMemoryStorageLimit = 1024 * 1024
	config.Store.Storage.MaxRecordSizeInBytes = 1048576
}

func TestMemstore_SetAndGet(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"
	invalidValue := map[int]int{1: 2}

	success, code := m.Set(key, invalidValue, 0)
	if success || code != entity.CRC_INVALID_DATATYPE {
		t.Errorf("expected invalid datatype err, got %v, %d", success, code)
	}

	success, code = m.Set(key, value, 0)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected successful set, got %v, %d", success, code)
	}

	record, code := m.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND || record.GetValue() != value {
		t.Errorf("expected value %v, got %v, code %d", value, record, code)
	}
}

func TestMemstore_Exists(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"

	m.Set(key, value, 1)

	exists, code := m.Exists(key)
	if !exists || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected record to exist, got %v, %d", exists, code)
	}

	time.Sleep(2 * time.Second)

	exists, code = m.Exists(key)
	if exists || code != entity.CRC_RECORD_EXPIRED {
		t.Errorf("expected record to not exist, got %v, %d", exists, code)
	}

	exists, code = m.Exists("nonExistentKey")
	if exists || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected record to not exist, got %v, %d", exists, code)
	}
}

func TestMemstore_Delete(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"

	m.Set(key, value, 0)

	deleted, code := m.Delete(key)
	if !deleted || code != entity.CRC_RECORD_DELETED {
		t.Errorf("expected record to be deleted, got %v, %d", deleted, code)
	}

	record, code := m.Get(key)
	if record != nil || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected record to not be found after deletion, got %v, %d", record, code)
	}
}

func TestMemstore_KeyExpired(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"
	ttl := int64(1)

	m.Set(key, value, ttl)

	record, code := m.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected record to be found before expiry, got %v, %d", record, code)
	}

	time.Sleep(2 * time.Second)

	record, code = m.Get(key)
	if record != nil || code != entity.CRC_RECORD_EXPIRED {
		t.Errorf("expected record to be expired, got %v, %d", record, code)
	}
}

func TestMemstore_Expire(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"
	ttl := int64(10)

	m.Set(key, value, ttl)
	success, code := m.Expire(key, 200)

	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected record to be updated with expiry, got code: %v, %d", success, code)
	}

	ttl, code = m.TTL(key)
	if ttl < 20 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected remaining TTL to be more than 20, got code: %v, %d", ttl, code)
	}
}

func TestMemstore_IncrDecr(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	initialValue := int64(10)

	m.Set(key, initialValue, 0)

	newValue, code := m.IncrDecrInteger(key, 5, true)
	if newValue != 15 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 15, got %d, %d", newValue, code)
	}

	newValue, code = m.IncrDecrInteger(key, 3, false)
	if newValue != 12 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 12, got %d, %d", newValue, code)
	}
}

func TestMemstore_Append(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	initialValue := "abcd"
	initialInvalidValue := 100

	m.Set(key, initialInvalidValue, 0)
	actualLen, code := m.Append(key, "_pqr")
	if actualLen != config.InvalidNumericValue || code != entity.CRC_INCR_INVALID_TYPE {
		t.Errorf("expected incr to fail, got %d, %d", actualLen, code)
	}

	m.Set(key, initialValue, 0)

	actualLen, code = m.Append(key, "_pqr")
	expectedLen := int64(len("abcd_pqr"))
	if actualLen != expectedLen || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value %d, got %d, %d", expectedLen, actualLen, code)
	}
}

func TestMemstore_MSet(t *testing.T) {
	SetUpMemstoreTests()
	m := CreateNewMemoryStore()

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	result, code := m.MSet(kvMap)
	if code != entity.CRC_MSET_COMPLETED {
		t.Errorf("MSet: expected %d, got %d", entity.CRC_MSET_COMPLETED, code)
	}

	for k, v := range result {
		if !v.(bool) {
			t.Errorf("MSet: expected %s key to be set, got %v", k, v)
		}
	}
}

func TestMemstore_MGet(t *testing.T) {
	SetUpMemstoreTests()
	m := CreateNewMemoryStore()

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	m.MSet(kvMap)
	result, code := m.MGet([]string{"key1", "key2", "key3"})

	if code != entity.CRC_MGET_COMPLETED {
		t.Errorf("MGet: expected %d, got %d", entity.CRC_MGET_COMPLETED, code)
	}

	r1 := result["key1"]
	r3 := result["key3"]

	if r1v, ok := r1.(map[string]interface{}); ok && r1v["Value"] != "value1" {
		t.Errorf("MGet: expected value1 for key1, got %v", r1v)
	}

	if r3v, ok := r3.(map[string]interface{}); ok && r3v["Value"] != nil {
		t.Errorf("MGet: expected nil for key3, got %v", r3v["Value"])
	}
}

func TestMemstore_MDelete(t *testing.T) {
	SetUpMemstoreTests()
	m := CreateNewMemoryStore()

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	m.MSet(kvMap)
	result, code := m.MDelete([]string{"key1", "key2"})
	if code != entity.CRC_MDEL_COMPLETED {
		t.Errorf("MDel: expected %d, got %d", entity.CRC_MDEL_COMPLETED, code)
	}

	for k, v := range result {
		if !v.(bool) {
			t.Errorf("MDel: expected %s key to be true, got %v", k, v)
		}
	}
}

func TestMemstore_TTL(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	key := "testKey"
	value := "testValue"
	ttl := int64(2)

	m.Set(key, value, ttl)

	remainingTTL, code := m.TTL(key)
	if remainingTTL <= 0 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected TTL greater than 0, got %d, %d", remainingTTL, code)
	}

	time.Sleep(3 * time.Second)

	remainingTTL, code = m.TTL(key)
	if remainingTTL != 0 || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected TTL to be 0 after expiry, got %d, %d", remainingTTL, code)
	}
}
func TestMemstore_GetStoreType(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	storeType := m.GetStoreType()

	if storeType != config.StorageEngineMemory {
		t.Errorf("expected store type %s, got %s", config.StorageEngineMemory, storeType)
	}
}
func TestMemstore_GetAllShards(t *testing.T) {
	SetUpMemstoreTests()

	m := CreateNewMemoryStore()
	shards := m.GetAllShards()

	if len(shards) != int(ShardCount) {
		t.Errorf("expected %d shards, got %d", ShardCount, len(shards))
	}

	for i, shard := range shards {
		if shard == nil {
			t.Errorf("expected shard at index %d to be initialized, got nil", i)
		}
	}
}
