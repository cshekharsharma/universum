package engine

import (
	"os"
	"universum/config"
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
	keyCount, err := PopulateRecordsFromSnapshot()

	if err != nil {
		logger.Get().Error("Translog replay failed, KeyOffset=%d, Err=%v", keyCount+1, err.Error())
		Shutdown()
	} else {
		logger.Get().Info("Translog replay done. Total %d keys replayed into DB", keyCount)
	}

	expiryJobExecutionFrequency = config.GetAutoRecordExpiryFrequency()
	snapshotJobExecutionFrequency = config.GetAutoSnapshotFrequency()

	// Trigger periodic jobs
	triggerPeriodicExpiryJob()
	triggerPeriodicSnapshotJob()
}

func Shutdown() {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
	StartInMemoryDBSnapshot(GetMemstore())
	os.Exit(0)
}
