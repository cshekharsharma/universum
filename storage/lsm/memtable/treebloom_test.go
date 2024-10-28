package memtable

import (
	"fmt"
	"testing"
	"time"
	"universum/config"
	"universum/entity"
)

func SetUpTBTests(t *testing.T) {
	tmpdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Storage.StorageEngine = config.StorageEngineLSM
	config.Store.Storage.MaxRecordSizeInBytes = 1048576
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeTB
	config.Store.Storage.LSM.WriteBufferSize = 1048576
	config.Store.Logging.LogFileDirectory = tmpdir
}

func TestTreeBloomMemTable_SetAndGet(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	invalidValue := map[int]int{1: 2}

	success, code := mt.Set(key, invalidValue, 0, entity.RecordStateActive)
	if success || code != entity.CRC_INVALID_DATATYPE {
		t.Errorf("expected invalid datatype error, got %v, %d", success, code)
	}

	success, code = mt.Set(key, value, 0, entity.RecordStateActive)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected successful set, got %v, %d", success, code)
	}

	record, code := mt.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND || record.(*entity.ScalarRecord).Value != value {
		t.Errorf("expected value %v, got %v, code %d", value, record, code)
	}
}

func TestTreeBloomMemTable_Exists(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	mt.Set(key, value, 1, entity.RecordStateActive)

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

func TestTreeBloomMemTable_Delete(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	mt.Set(key, value, 0, entity.RecordStateActive)

	deleted, code := mt.Delete(key)
	if !deleted || code != entity.CRC_RECORD_DELETED {
		t.Errorf("expected record to be deleted, got %v, %d", deleted, code)
	}

	record, code := mt.Get(key)
	if code != entity.CRC_RECORD_TOMBSTONED {
		t.Errorf("expected record to be tombstoned, got %v, %d", record, code)
	}
}

func TestTreeBloomMemTable_SizeManagement(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"

	initialSize := mt.GetSize()

	mt.Set(key, value, 0, entity.RecordStateActive)
	updatedSize := mt.GetSize()
	if updatedSize <= initialSize {
		t.Errorf("expected size to increase after set, got %d", updatedSize)
	}

	mt.Delete(key)
	finalSize := mt.GetSize()
	if finalSize > updatedSize {
		t.Errorf("expected final size to be smaller than updated size, got %d", finalSize)
	}
}

func TestTreeBloomMemTable_KeyExpired(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(1)

	mt.Set(key, value, ttl, entity.RecordStateActive)

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

func TestTreeBloomMemTable_Expire(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(10)

	mt.Set(key, value, ttl, entity.RecordStateActive)
	success, code := mt.Expire(key, 200)

	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected record to be updated with expiry, got code: %v, %d", success, code)
	}

	ttl, code = mt.TTL(key)
	if ttl < 20 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected remaining TTL to be more than 20, got code: %v, %d", ttl, code)
	}
}

func TestTreeBloomMemTable_IncrDecr(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	initialValue := int64(10)

	mt.Set(key, initialValue, 0, entity.RecordStateActive)

	newValue, code := mt.IncrDecrInteger(key, 5, true)
	if newValue != 15 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 15, got %d, %d", newValue, code)
	}

	newValue, code = mt.IncrDecrInteger(key, 3, false)
	if newValue != 12 || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value 12, got %d, %d", newValue, code)
	}
}

func TestTreeBloomMemTable_Append(t *testing.T) {
	SetUpTBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	initialValue := "abcd"
	initialInvalidValue := 100

	mt.Set(key, initialInvalidValue, 0, entity.RecordStateActive)
	actualLen, code := mt.Append(key, "_pqr")
	if actualLen != config.InvalidNumericValue || code != entity.CRC_INCR_INVALID_TYPE {
		t.Errorf("expected append to fail, got %d, %d", actualLen, code)
	}

	mt.Set(key, initialValue, 0, entity.RecordStateActive)

	actualLen, code = mt.Append(key, "_pqr")
	expectedLen := int64(len("abcd_pqr"))
	if actualLen != expectedLen || code != entity.CRC_RECORD_UPDATED {
		t.Errorf("expected new value %d, got %d, %d", expectedLen, actualLen, code)
	}
}

func TestTreeBloomMemTable_MSet(t *testing.T) {
	SetUpTBTests(t)
	tbMem := NewTreeBloomMemTable(100, 0.01)

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	result, code := tbMem.MSet(kvMap)
	if code != entity.CRC_MSET_COMPLETED {
		t.Errorf("MSet: expected %d, got %d", entity.CRC_MSET_COMPLETED, code)
	}

	for k, v := range result {
		if !v.(bool) {
			t.Errorf("MSet: expected %s key to be set, got %v", k, v)
		}
	}
}

func TestTreeBloomMemTable_MGet(t *testing.T) {
	SetUpTBTests(t)
	tbMem := NewTreeBloomMemTable(100, 0.01)

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	tbMem.MSet(kvMap)
	result, code := tbMem.MGet([]string{"key1", "key2", "key3"})

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

func TestTreeBloomMemTable_MDelete(t *testing.T) {
	SetUpLBTests(t)
	lbMem := NewTreeBloomMemTable(100, 0.01)

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	lbMem.MSet(kvMap)
	result, code := lbMem.MDelete([]string{"key1", "key2"})
	if code != entity.CRC_MDEL_COMPLETED {
		t.Errorf("MDel: expected %d, got %d", entity.CRC_MDEL_COMPLETED, code)
	}

	for k, v := range result {
		if !v.(bool) {
			t.Errorf("MDel: expected %s key to be true, got %v", k, v)
		}
	}
}

func TestTreeBloomMemTable_TTL(t *testing.T) {
	SetUpLBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	key := "testKey"
	value := "testValue"
	ttl := int64(1)

	mt.Set(key, value, ttl, entity.RecordStateActive)

	remainingTTL, code := mt.TTL(key)
	if remainingTTL <= 0 || code != entity.CRC_RECORD_FOUND {
		t.Errorf("expected TTL greater than 0, got %d, %d", remainingTTL, code)
	}

	time.Sleep(2 * time.Second)

	remainingTTL, code = mt.TTL(key)
	if remainingTTL != 0 || code != entity.CRC_RECORD_NOT_FOUND {
		t.Errorf("expected TTL to be 0 after expiry, got %d, %d", remainingTTL, code)
	}
}

func TestTreeBloomMemTable_IsFull(t *testing.T) {
	SetUpLBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	isfull := mt.IsFull()

	if isfull == true {
		t.Errorf("expected false, got %v", isfull)
	}
}

func TestTreeBloomMemTable_GetCount(t *testing.T) {
	SetUpLBTests(t)

	mt := NewTreeBloomMemTable(100, 0.01)
	count := mt.GetCount()

	if count != 0 {
		t.Errorf("expected count to be 0, got %d", count)
	}

	mt.Set("key1", "value1", 0, entity.RecordStateActive)
	mt.Set("key2", "value2", 0, entity.RecordStateTombstoned)
	count = mt.GetCount()

	if count != 2 {
		t.Errorf("expected count to be 2, got %d", count)
	}
}

func TestTreeBloomMemTable_Truncate(t *testing.T) {
	SetUpTBTests(t)
	config.Store.Storage.LSM.WriteBufferSize = 100

	FlusherChan = make(chan MemTable, 2)
	WALRotaterChan = make(chan int64, 2)

	tbMem := NewTreeBloomMemTable(100, 0.01)

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
	}

	tbMem.MSet(kvMap)

	if tbMem.GetCount() != 1 {
		t.Errorf("Expected memtable count to be 1 after truncation, found: %d", tbMem.GetCount())
	}

	select {
	case item := <-FlusherChan:
		var backupMemtable *TreeBloomMemTable
		backupMemtable = item.(*TreeBloomMemTable)

		if backupMemtable.GetCount() != 3 {
			t.Error("Expected flushed memtable to have 3 records")
		}
	default:
		t.Error("Expected an entry in FlusherChan for table flushing")
	}
}

func TestTreeBloomMemTable_GetAll(t *testing.T) {
	SetUpTBTests(t)

	FlusherChan = make(chan MemTable, 1)
	tbMem := NewTreeBloomMemTable(100, 0.01)

	kvMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	tbMem.MSet(kvMap)
	allRecords := tbMem.GetAll()

	if len(allRecords) != 3 {
		t.Errorf("Expected 3 records to be returned, got %d", len(allRecords))
	}

	for i := 0; i < 3; i++ {
		recordKV := allRecords[i]
		key := recordKV.Key
		record, ok := recordKV.Record.(*entity.ScalarRecord)
		if !ok {
			t.Errorf("Expected record to be of type ScalarRecord, got %v", allRecords[i])
		}

		if key != fmt.Sprintf("key%d", i+1) {
			t.Errorf("Expected key %v, got %v", fmt.Sprintf("key%d", i+1), key)
		}

		if record.Value != kvMap[fmt.Sprintf("key%d", i+1)] {
			t.Errorf("Expected value %v, got %v", kvMap[fmt.Sprintf("key%d", i+1)], record.Value)
		}
	}
}
