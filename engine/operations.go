package engine

import (
	"os"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
)

var memstore *storage.MemoryStore

func GetMemstore() *storage.MemoryStore {
	return memstore
}

func Startup() {
	// Initiaise the in-memory store
	memstore = storage.CreateNewMemoryStore()
	memstore.Initialize()

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
	if exitcode != entity.ExitCodeStartupFailure {
		StartInMemoryDBSnapshot(GetMemstore())
	}

	os.Exit(exitcode)
}
