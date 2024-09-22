package memtable

import (
	"testing"
	"time"
	"universum/config"
	"universum/entity"
)

func SetUpTests() {
	cfg := `
[storage]
StorageEngine="LSM"
MaxRecordSizeInBytes=1048576

[storage.lsm]
MemtableStorageType="LB"
`
	config.LoadFromString(cfg)
}

func TestListBloomMemTable_SetAndGet(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	invalidValue := map[int]int{1: 2}

	success, code := mt.Set(key, invalidValue, 0)
	if success || code != entity.CRC_INVALID_DATATYPE {
		t.Errorf("expected invalid datatype err, got %v, %d", success, code)
	}

	success, code = mt.Set(key, value, 0)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected successful set, got %v, %d", success, code)
	}

	record, code := mt.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND || record.(*entity.ScalarRecord).Value != value {
		t.Errorf("expected value %v, got %v, code %d", value, record, code)
	}
}

func TestListBloomMemTable_Exists(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	mt.Set(key, value, 1)

	exists, code := mt.Exists(key)
	if !exists || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected record to exist, got %v, %d", exists, code)
	}

	time.Sleep(2 * time.Second)

	exists, code = mt.Exists(key)
	if exists || code != entity.CRC_RECORD_EXPIRED {
		t.Errorf("expected record to not exist, got %v, %d", exists, code)
	}

	exists, code = mt.Exists("nonExistentKey")
	if exists || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected record to not exist, got %v, %d", exists, code)
	}
}

func TestListBloomMemTable_Delete(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	mt.Set(key, value, 0)

	deleted, code := mt.Delete(key)
	if !deleted || code != entity.CRC_RECORD_DELETED {
		t.Errorf("expected record to be deleted, got %v, %d", deleted, code)
	}

	record, code := mt.Get(key)
	if record != nil || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected record to not be found after deletion, got %v, %d", record, code)
	}
}

func TestListBloomMemTable_SizeManagement(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	initialSize := mt.GetSize()

	mt.Set(key, value, 0)
	updatedSize := mt.GetSize()
	if updatedSize <= initialSize {
		t.Errorf("expected size to increase after set, got %d", updatedSize)
	}

	mt.Delete(key)
	finalSize := mt.GetSize()
	if finalSize != initialSize {
		t.Errorf("expected size to return to initial, got %d", finalSize)
	}
}

func TestListBloomMemTable_KeyExpired(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(1)

	mt.Set(key, value, ttl)

	record, code := mt.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected record to be found before expiry, got %v, %d", record, code)
	}

	time.Sleep(2 * time.Second)

	record, code = mt.Get(key)
	if record != nil || code != entity.CRC_RECORD_EXPIRED {
		t.Errorf("expected record to be expired, got %v, %d", record, code)
	}
}

func TestListBloomMemTable_Expire(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(10)

	mt.Set(key, value, ttl)
	success, code := mt.Expire(key, 200)

	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected record to be updated with expiry, got code: %v, %d", success, code)
	}

	ttl, code = mt.TTL(key)
	if ttl < 20 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected remaining TTL to be more than 20, got code: %v, %d", ttl, code)
	}
}

func TestListBloomMemTable_IncrDecr(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	initialValue := int64(10)

	mt.Set(key, initialValue, 0)

	newValue, code := mt.IncrDecrInteger(key, 5, true)
	if newValue != 15 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 15, got %d, %d", newValue, code)
	}

	newValue, code = mt.IncrDecrInteger(key, 3, false)
	if newValue != 12 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 12, got %d, %d", newValue, code)
	}
}

func TestListBloomMemTable_Append(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	initialValue := "abcd"
	initialInvalidValue := 100

	mt.Set(key, initialInvalidValue, 0)
	actualLen, code := mt.Append(key, "_pqr")
	if actualLen != config.InvalidNumericValue || code != entity.CRC_INCR_INVALID_TYPE {
		t.Errorf("expected incr to fail, got %d, %d", actualLen, code)
	}

	mt.Set(key, initialValue, 0)

	actualLen, code = mt.Append(key, "_pqr")
	expectedLen := int64(len("abcd_pqr"))
	if actualLen != expectedLen || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value %d, got %d, %d", expectedLen, actualLen, code)
	}
}

func TestListBloomMemTable_TTL(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(2)

	mt.Set(key, value, ttl)

	remainingTTL, code := mt.TTL(key)
	if remainingTTL <= 0 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected TTL greater than 0, got %d, %d", remainingTTL, code)
	}

	time.Sleep(3 * time.Second)

	remainingTTL, code = mt.TTL(key)
	if remainingTTL != 0 || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected TTL to be 0 after expiry, got %d, %d", remainingTTL, code)
	}
}

func TestIsFull(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	isfull := mt.IsFull()

	if isfull == true {
		t.Errorf("expected false, got %v", isfull)
	}
}

func TestGetRecordCount(t *testing.T) {
	SetUpTests()

	mt := NewListBloomMemTable(100, 0.01)
	count := mt.GetRecordCount()

	if count != 0 {
		t.Errorf("expected count to be 0, got %d", count)
	}

	mt.Set("key1", "value1", 0)
	count = mt.GetRecordCount()

	if count != 1 {
		t.Errorf("expected count to be 1, got %d", count)
	}
}
