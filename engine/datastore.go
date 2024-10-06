package engine

import (
	"strings"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
	"universum/storage/lsm"
	"universum/storage/memory"
)

var datastore storage.DataStore

var _allStores = make(map[string]storage.DataStore)

func getDataStore(id string) storage.DataStore {
	id = strings.ToUpper(id) // to avoid casing typos

	switch id {
	case config.StorageEngineMemory:
		if _, ok := _allStores[id]; !ok {
			_allStores[config.StorageEngineMemory] = memory.CreateNewMemoryStore()
		}
		return _allStores[config.StorageEngineMemory]

	case config.StorageEngineLSM:
		if _, ok := _allStores[id]; !ok {
			memtableType := config.Store.Storage.LSM.MemtableStorageType
			_allStores[config.StorageEngineMemory] = lsm.CreateNewLSMStore(memtableType)
		}
		return _allStores[config.StorageEngineMemory]

	default:
		logger.Get().Error("GetDataStore: unknown storage engine `%s` requested, shutting down.", id)
		Shutdown(entity.ExitCodeStartupFailure)
		return nil
	}
}

func getSnapshotService(id string) storage.SnapshotService {
	id = strings.ToUpper(id) // to avoid casing typos

	switch id {
	case config.StorageEngineMemory:
		return new(memory.MemoryStoreSnapshotService)

	default:
		logger.Get().Error("GetSnapshotService: unknown storage engine `%s` requested, shutting down.", id)
		Shutdown(entity.ExitCodeStartupFailure)
		return nil
	}
}
