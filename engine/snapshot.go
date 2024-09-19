package engine

import (
	"fmt"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
	"universum/storage/memory"
	"universum/utils"
)

var snapshotJobLastExecutedAt time.Time
var snapshotJobExecutionFrequency time.Duration

type databaseSnapshotWorker struct {
	ExecutionErr error
}

func (w *databaseSnapshotWorker) startDatabaseSnapshot(snapshotChan chan<- databaseSnapshotWorker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				w.ExecutionErr = err
			} else {
				w.ExecutionErr = fmt.Errorf("panic happened with %v", r)
			}
		} else {
			w.ExecutionErr = err
		}

		// emit message to the worker channel that worker is dying.
		snapshotChan <- *w
	}()

	snapshotJobLastExecutedAt = time.Now()

	for {
		nextScheduledTime := snapshotJobLastExecutedAt.Add(expiryJobExecutionFrequency)

		if nextScheduledTime.Compare(time.Now()) < 1 {
			StartDataBaseSnapshot(getDataStore(config.GetStorageEngineType()))
			snapshotJobLastExecutedAt = time.Now()
		}

		time.Sleep(snapshotJobExecutionFrequency)
	}
}

func StartDataBaseSnapshot(store storage.DataStore) error {
	var recordCount int64 = 0
	var ssSizeInBytes int64 = 0
	var snapshotStartTime int64 = time.Now().UnixMilli()
	var err error

	snapshotservice := getSnapshotService(config.GetStorageEngineType())
	recordCount, ssSizeInBytes, err = snapshotservice.StartDataBaseSnapshot(store.(*memory.MemoryStore))

	DatabaseInfoStats.Persistence.LastSnapshotTakenAt = utils.GetCurrentReadableTime()
	DatabaseInfoStats.Persistence.LastSnapshotLatency = fmt.Sprintf("%dms", time.Now().UnixMilli()-snapshotStartTime)
	DatabaseInfoStats.Persistence.SnapshotSizeInBytes = ssSizeInBytes
	DatabaseInfoStats.Keyspace.TotalKeyCount = recordCount

	return err
}

func ReplayDBRecordsFromSnapshot(datastore storage.DataStore) {
	replayStartTime := time.Now().UnixMilli()

	snapshotservice := getSnapshotService(config.GetStorageEngineType())
	keyCount, err := snapshotservice.ReplayDBRecordsFromSnapshot(datastore)

	replayLatency := fmt.Sprintf("%d ms", time.Now().UnixMilli()-replayStartTime)

	DatabaseInfoStats.Persistence.LastSnapshotReplayLatency = replayLatency
	DatabaseInfoStats.Persistence.TotalKeysReplayed = keyCount
	DatabaseInfoStats.Persistence.LastSnapshotReplayedAt = utils.GetCurrentReadableTime()

	if err != nil {
		logger.Get().Error("Snapshot replay failed, KeyOffset=%d, Err=%v", keyCount+1, err.Error())
		Shutdown(entity.ExitCodeStartupFailure)
	} else {
		logger.Get().Info("Snapshot replay done. Total %d keys replayed into DB", keyCount)
	}
}
