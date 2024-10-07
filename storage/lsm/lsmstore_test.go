package lsm

import (
	"testing"
	"universum/config"
	"universum/entity"
)

func setupTestStore(t *testing.T) {
	tempdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Logging.LogFileDirectory = tempdir
	config.Store.Storage.MaxRecordSizeInBytes = 1024
	config.Store.Storage.StorageEngine = config.StorageEngineLSM

	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.BlockCompressionAlgo = config.CompressionAlgoLZ4
	config.Store.Storage.LSM.DataStorageDirectory = tempdir
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeLB
	config.Store.Storage.LSM.MaxMemtableRecords = 100
	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false
	config.Store.Storage.LSM.WriteAheadLogDirectory = tempdir
	config.Store.Storage.LSM.WriteAheadLogBufferSize = 1024
	config.Store.Storage.LSM.WriteAheadLogFrequency = 2
	config.Store.Storage.LSM.WriteBlockSize = 1024
}

func TestNewLSMStore(t *testing.T) {
	setupTestStore(t)
	store := CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)

	if store == nil {
		t.Fatal("Expected LSMStore instance, got nil")
	}

	if store.memTable == nil {
		t.Fatal("Expected MemTable to be initialized, got nil")
	}

	if len(store.sstables) != 0 {
		t.Fatalf("Expected SSTables to be empty, got %d", len(store.sstables))
	}
}

func TestLSMStoreInitialize(t *testing.T) {
	setupTestStore(t)
	store := CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)

	err := store.Initialize()
	if err != nil {
		t.Fatalf("Initialization failed with error: %v", err)
	}

	if store.walWriter == nil {
		t.Fatal("Expected WALWriter to be initialized, got nil")
	}
}

func TestLSMStoreSetAndExists(t *testing.T) {
	setupTestStore(t)
	store := CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)
	store.Initialize()

	success, code := store.Set("key1", "value1", 100)
	if !success || code != entity.CRC_RECORD_UPDATED {
		t.Fatalf("Expected success and CRC_RECORD_UPDATED, got success=%v and code=%d", success, code)
	}

	exists, _ := store.Exists("key1")
	if !exists {
		t.Fatal("Expected key 'key1' to exist after Set, but it does not")
	}
}

func TestLSMStoreGet(t *testing.T) {
	setupTestStore(t)
	store := CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)
	store.Initialize()

	store.Set("key1", "value1", 100)

	record, code := store.Get("key1")
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Fatalf("Expected record to be found for 'key1', but got code=%d", code)
	}

	if record.(*entity.ScalarRecord).Value != "value1" {
		t.Fatalf("Expected value 'value1', but got '%v'", record.(*entity.ScalarRecord).Value)
	}
}

func TestLSMStoreExistsWhenKeyDoesNotExist(t *testing.T) {
	setupTestStore(t)
	store := CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)
	store.Initialize()

	exists, _ := store.Exists("nonexistent_key")
	if exists {
		t.Fatal("Expected key 'nonexistent_key' to not exist, but it does")
	}
}
