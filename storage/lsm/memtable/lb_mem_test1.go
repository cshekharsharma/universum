package memtable

import (
	"testing"
	"time"
	"universum/entity"
)

func TestListBloomMemTable_SetAndGet(t *testing.T) {
	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	success, code := mt.Set(key, value, 0)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected successful set, got %v, %d", success, code)
	}

	record, code := mt.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND || record.(*entity.ScalarRecord).Value != value {
		t.Errorf("expected value %v, got %v, code %d", value, record, code)
	}
}

func TestListBloomMemTable_Exists(t *testing.T) {
	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	mt.Set(key, value, 0)

	exists, code := mt.Exists(key)
	if !exists || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected record to exist, got %v, %d", exists, code)
	}

	exists, code = mt.Exists("nonExistentKey")
	if exists || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected record to not exist, got %v, %d", exists, code)
	}
}

func TestListBloomMemTable_Delete(t *testing.T) {
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

func TestListBloomMemTable_Expire(t *testing.T) {
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

func TestListBloomMemTable_IncrDecr(t *testing.T) {
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

func TestListBloomMemTable_TTL(t *testing.T) {
	mt := NewListBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(5)

	mt.Set(key, value, ttl)

	remainingTTL, code := mt.TTL(key)
	if remainingTTL <= 0 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected TTL greater than 0, got %d, %d", remainingTTL, code)
	}

	time.Sleep(6 * time.Second)

	remainingTTL, code = mt.TTL(key)
	if remainingTTL != 0 || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected TTL to be 0 after expiry, got %d, %d", remainingTTL, code)
	}
}
