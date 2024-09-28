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
	case config.StorageTypeMemory:
		if _, ok := _allStores[id]; !ok {
			_allStores[config.StorageTypeMemory] = memory.CreateNewMemoryStore()
		}
		return _allStores[config.StorageTypeMemory]

	case config.StorageTypeLSM:
		if _, ok := _allStores[id]; !ok {
			_allStores[config.StorageTypeMemory] = lsm.CreateNewLSMStore(config.Store.Storage.LSM.MemtableStorageType)
		}
		return _allStores[config.StorageTypeMemory]

	default:
		logger.Get().Error("GetDataStore: unknown storage engine `%s` requested, shutting down.", id)
		Shutdown(entity.ExitCodeStartupFailure)
		return nil
	}
}

func getSnapshotService(id string) storage.SnapshotService {
	id = strings.ToUpper(id) // to avoid casing typos

	switch id {
	case config.StorageTypeMemory:
		return new(memory.MemoryStoreSnapshotService)

	default:
		logger.Get().Error("GetSnapshotService: unknown storage engine `%s` requested, shutting down.", id)
		Shutdown(entity.ExitCodeStartupFailure)
		return nil
	}
}
