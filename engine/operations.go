package engine

import (
	"os"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
	"universum/utils"
)

var store storage.DataStore

func GetStore() storage.DataStore {
	store, err := storage.GetStore(config.GetStorageEngine())
	if err != nil {
		logger.Get().Error("Error initialising storage engine: %v", err)
		Shutdown(entity.ExitCodeStartupFailure)
	}
	return store
}

func Startup() {
	// Initiaise the data store
	store = GetStore()
	store.Initialize()

	// Replay all commands from translog into the database
	aofFile := config.GetTransactionLogFilePath()
	if _, err := os.Stat(aofFile); os.IsNotExist(err) {
		_, err := os.Create(aofFile)
		if err != nil {
			logger.Get().Error("Error creating AOF file, shutting down: %v", err)
			Shutdown(entity.ExitCodeStartupFailure)
		}
	}

	keyCount, err := ReplayDBRecordsFromSnapshot()

	if err != nil {
		logger.Get().Error("Translog replay failed, KeyOffset=%d, Err=%v", keyCount+1, err.Error())
		Shutdown(entity.ExitCodeStartupFailure)
	} else {
		logger.Get().Info("Translog replay done. Total %d keys replayed into DB", keyCount)
	}

	expiryJobExecutionFrequency = config.GetAutoRecordExpiryFrequency()
	snapshotJobExecutionFrequency = config.GetAutoSnapshotFrequency()

	// Trigger periodic jobs
	triggerPeriodicExpiryJob()
	triggerPeriodicSnapshotJob()
	triggerPeriodicEvictionJob()
}

func Shutdown(exitcode int) {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
	nonSnapshotErrs := []int{
		entity.ExitCodeStartupFailure,
		entity.ExitCodeSocketError,
	}

	shouldSkipSnapshot, _ := utils.ExistsInList(exitcode, nonSnapshotErrs)

	if !shouldSkipSnapshot {
		StartDataBaseSnapshot(GetStore())
	}

	os.Exit(exitcode)
}
