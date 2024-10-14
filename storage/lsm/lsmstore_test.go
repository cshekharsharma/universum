package lsm

import (
	"fmt"
	"testing"
	"time"
	"universum/config"
	"universum/entity"
)

func setupTestStore(t *testing.T) *LSMStore {
	tempdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Logging.LogFileDirectory = tempdir
	config.Store.Storage.MaxRecordSizeInBytes = 1024
	config.Store.Storage.StorageEngine = config.StorageEngineLSM

	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.BlockCompressionAlgo = config.CompressionAlgoLZ4
	config.Store.Storage.LSM.DataStorageDirectory = tempdir
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeLB
	config.Store.Storage.LSM.BloomFilterMaxRecords = 100
	config.Store.Storage.LSM.WriteBufferSize = 1024 * 1024
	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false
	config.Store.Storage.LSM.WriteAheadLogDirectory = tempdir
	config.Store.Storage.LSM.WriteAheadLogBufferSize = 1024
	config.Store.Storage.LSM.WriteAheadLogFrequency = 10
	config.Store.Storage.LSM.WriteBlockSize = 1024

	store := CreateNewLSMStore(config.MemtableStorageTypeLB)
	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	return store
}

func TestLSMStoreSetGetExists(t *testing.T) {
	store := setupTestStore(t)
	key := "test-key"
	value := "test-value"

	success, code := store.Set(key, value, 60)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Set operation failed before flush")
	}

	exists, code := store.Exists(key)
	if !exists || code != entity.CRC_RECORD_FOUND {
		t.Fatalf("Exists operation failed before flush")
	}

	record, code := store.Get(key)
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Fatalf("Get operation failed before flush")
	}

	if record.(*entity.ScalarRecord).Value != value {
		t.Fatalf("Record value mismatch before flush")
	}
}

func TestLSMStoreSetGetExistsAfterFlush(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("test-key-flush-%d", i)
		value := fmt.Sprintf("test-key-flush-%d", i)

		success, code := store.Set(key, value, 6000)
		if !success || code != entity.CRC_RECORD_UPDATED {
			t.Fatalf("Set operation failed before flush")
		}
	}

	store.memTable.Truncate()
	time.Sleep(2 * time.Second)

	keysToCheck := []string{"test-key-flush-0", "test-key-flush-50", "test-key-flush-99"}

	for i := 0; i < len(keysToCheck); i++ {
		key := keysToCheck[i]

		exists, code := store.Exists(key)
		if !exists || code != entity.CRC_RECORD_FOUND {
			t.Fatalf("Exists operation failed after flush")
		}

		record, code := store.Get(key)
		if record == nil || code != entity.CRC_RECORD_FOUND {
			t.Fatalf("Get operation failed after flush")
		}

		if record.(*entity.ScalarRecord).Value != key {
			t.Fatalf("Record value mismatch after flush")
		}
	}
}

func TestDeleteOperation(t *testing.T) {
	store := setupTestStore(t)

	key := "test-delete"
	value := "test-value-delete"

	success, code := store.Set(key, value, 60)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Set operation failed")
	}

	deleted, code := store.Delete(key)
	if !deleted || code != entity.CRC_RECORD_DELETED {
		t.Fatalf("Delete operation failed before flush")
	}

	exists, _ := store.Exists(key)
	if exists {
		t.Fatalf("Exists should return false for deleted record before flush")
	}
}

func TestDeleteOperationAfterFlush(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("test-key-flush-%d", i)
		value := fmt.Sprintf("test-key-flush-%d", i)

		success, code := store.Set(key, value, 6000)
		if !success || code != entity.CRC_RECORD_UPDATED {
			t.Fatalf("Set operation failed before flush")
		}
	}

	store.memTable.Truncate()
	time.Sleep(2 * time.Second)

	keysToCheck := []string{"test-key-flush-0", "test-key-flush-50", "test-key-flush-99"}

	for i := 0; i < len(keysToCheck); i++ {
		key := keysToCheck[i]
		deleted, code := store.Delete(key)
		if !deleted || code != entity.CRC_RECORD_DELETED {
			t.Fatalf("Delete operation failed after flush")
		}

		exists, _ := store.Exists(key)
		if exists {
			t.Fatalf("Exists should return false for deleted record after flush")
		}
	}
}

func TestIncrDecrOperation(t *testing.T) {
	store := setupTestStore(t)

	key := "test-incr-decr"
	initialValue := int64(100)

	success, _ := store.Set(key, initialValue, 60)
	if !success {
		t.Fatalf("Set operation failed")
	}

	newValue, _ := store.IncrDecrInteger(key, 10, true)
	if newValue != 110 {
		t.Fatalf("Increment operation failed, expected 110 but got %d", newValue)
	}

	newValue, _ = store.IncrDecrInteger(key, 20, false)
	if newValue != 90 {
		t.Fatalf("Decrement operation failed, expected 90 but got %d", newValue)
	}
}

func TestAppendOperation(t *testing.T) {
	store := setupTestStore(t)

	key := "test-append"
	initialValue := "hello"

	success, _ := store.Set(key, initialValue, 60)
	if !success {
		t.Fatalf("Set operation failed")
	}

	newLength, _ := store.Append(key, " world")
	if newLength != 11 {
		t.Fatalf("Append operation failed, expected length 11 but got %d", newLength)
	}

	record, _ := store.Get(key)
	if record.(*entity.ScalarRecord).Value != "hello world" {
		t.Fatalf("Append operation failed, expected 'hello world' but got %v", record.(*entity.ScalarRecord).Value)
	}
}

func TestTTL(t *testing.T) {
	store := setupTestStore(t)

	key := "test-key-ttl"
	value := "test-value-ttl"
	ttl := int64(10)

	success, code := store.Set(key, value, ttl)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Set operation with TTL failed")
	}

	actualTTL, code := store.TTL(key)
	if actualTTL == 0 || code != entity.CRC_RECORD_FOUND {
		t.Fatalf("Failed to retrieve TTL for key")
	}

	if actualTTL > ttl || actualTTL < (ttl-2) {
		t.Fatalf("TTL mismatch: expected ~%d, got %d", ttl, actualTTL)
	}
}

func TestExpire(t *testing.T) {
	store := setupTestStore(t)

	key := "test-key-expire"
	value := "test-value-expire"
	ttl := int64(60)

	success, code := store.Set(key, value, ttl)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Set operation with TTL failed")
	}

	newTTL := int64(30)
	success, code = store.Expire(key, newTTL)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Expire operation failed")
	}

	actualTTL, _ := store.TTL(key)
	if actualTTL != newTTL {
		t.Fatalf("Expire operation mismatch: expected TTL %d, got %d", newTTL, actualTTL)
	}
}

func TestMSetMGetMDeleteOperations(t *testing.T) {
	store := setupTestStore(t)
	kvMap := map[string]interface{}{}

	for i := 0; i < 100; i++ {
		kvMap[fmt.Sprintf("test-key-%d", i)] = fmt.Sprintf("test-value-%d", i)
	}

	_, code := store.MSet(kvMap)
	if code != entity.CRC_MSET_COMPLETED {
		t.Fatalf("MSet operation failed")
	}

	store.memTable.Truncate()
	time.Sleep(2 * time.Second)
	store.Initialize()

	keys := []string{"test-key-flush-0", "test-key-flush-50", "test-key-flush-99"}
	resultMap, code := store.MGet(keys)
	if code != entity.CRC_MGET_COMPLETED {
		t.Fatalf("MGet operation failed")
	}

	for _, key := range keys {
		val := resultMap[key].(map[string]interface{})["Value"]
		if val != kvMap[key] {
			t.Fatalf("MGet value mismatch for key: %s", key)
		}
	}

	deleteKeys := []string{"test-key-flush-1", "test-key-flush-51", "test-key-flush-98"}
	resultMap, code = store.MDelete(deleteKeys)
	if code != entity.CRC_MDEL_COMPLETED {
		t.Fatalf("MDelete operation failed")
	}

	for r := range resultMap {
		if !resultMap[r].(bool) {
			t.Fatalf("MDelete failed for key: %s", r)
		}
	}
}

func TestClose(t *testing.T) {
	store := setupTestStore(t)

	err := store.walWriter.AddToWALBuffer("key", "value", 0, entity.RecordStateActive)
	if err != nil {
		t.Fatalf("Failed to create dummy WAL file: %v", err)
	}

	err = store.Close()
	if err != nil {
		t.Fatalf("Failed to close the store: %v", err)
	}
}
