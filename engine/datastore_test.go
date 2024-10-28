package engine

import (
	"reflect"
	"testing"
	"universum/config"
	"universum/storage"
	"universum/storage/lsm"
	"universum/storage/memory"
)

func setupEngineTests() {
	_allStores = make(map[string]storage.DataStore)
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.MemtableStorageType = "TB"
}

func TestGetDataStoreMemory(t *testing.T) {
	setupEngineTests()
	datastore := getDataStore(config.StorageEngineMemory)

	if datastore == nil {
		t.Fatal("Expected a valid MemoryStore instance, got nil")
	}

	if _, ok := datastore.(*memory.MemoryStore); !ok {
		t.Fatalf("Expected datastore of type *memory.MemoryStore, got %T", datastore)
	}

	datastore2 := getDataStore(config.StorageEngineMemory)
	if datastore != datastore2 {
		t.Error("Expected getDataStore to return the same instance on repeated calls for MEMORY engine")
	}
}

func TestGetDataStoreLSM(t *testing.T) {
	setupEngineTests()
	datastore := getDataStore(config.StorageEngineLSM)

	if datastore == nil {
		t.Fatal("Expected a valid LSMStore instance, got nil")
	}

	if _, ok := datastore.(*lsm.LSMStore); !ok {
		t.Fatalf("Expected datastore of type *lsm.LSMStore, got %T", datastore)
	}

	datastore2 := getDataStore(config.StorageEngineLSM)
	if !reflect.DeepEqual(datastore, datastore2) {
		t.Error("Expected getDataStore to return the same instance on repeated calls for LSM engine")
	}
}

func TestGetSnapshotServiceMemory(t *testing.T) {
	setupEngineTests()
	snapshotService := getSnapshotService(config.StorageEngineMemory)

	if snapshotService == nil {
		t.Fatal("Expected a valid MemoryStoreSnapshotService instance, got nil")
	}

	if _, ok := snapshotService.(*memory.MemoryStoreSnapshotService); !ok {
		t.Fatalf("Expected snapshot service of type *memory.MemoryStoreSnapshotService, got %T", snapshotService)
	}
}

func TestGetSnapshotServiceLSM(t *testing.T) {
	setupEngineTests()
	snapshotService := getSnapshotService(config.StorageEngineLSM)

	if snapshotService == nil {
		t.Fatal("Expected a valid LSMStoreSnapshotService instance, got nil")
	}

	if _, ok := snapshotService.(*lsm.LSMStoreSnapshotService); !ok {
		t.Fatalf("Expected snapshot service of type *lsm.LSMStoreSnapshotService, got %T", snapshotService)
	}
}
