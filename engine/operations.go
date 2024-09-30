package engine

import (
	"os"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/utils"
)

func Startup() {
	// Initiaise the data store
	datastore = getDataStore(config.Store.Storage.StorageEngine)
	err := datastore.Initialize()

	if err != nil {
		logger.Get().Fatal("Application startup failed: %v", err)
		Shutdown(entity.ExitCodeStartupFailure)
	}

	RestoreDatabaseSnapshot(datastore)

	expiryJobExecutionFrequency = time.Duration(config.Store.Eviction.AutoRecordExpiryFrequency) * time.Second
	snapshotJobExecutionFrequency = time.Duration(config.Store.Storage.Memory.AutoSnapshotFrequency) * time.Second

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
		StartDatabaseSnapshot(getDataStore(config.Store.Storage.StorageEngine))
	}

	os.Exit(exitcode)
}
