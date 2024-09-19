package engine

import (
	"os"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/utils"
)

func Startup() {
	// Initiaise the data store
	datastore = getDataStore(config.GetStorageEngineType())
	datastore.Initialize()

	// Replay all commands from translog into the database
	aofFile := config.GetTransactionLogFilePath()
	if _, err := os.Stat(aofFile); os.IsNotExist(err) {
		_, err := os.Create(aofFile)
		if err != nil {
			logger.Get().Error("Error creating AOF file, shutting down: %v", err)
			Shutdown(entity.ExitCodeStartupFailure)
		}
	}

	ReplayDBRecordsFromSnapshot(datastore)

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
		StartDataBaseSnapshot(getDataStore(config.GetStorageEngineType()))
	}

	os.Exit(exitcode)
}
