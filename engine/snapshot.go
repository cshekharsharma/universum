package engine

import (
	"fmt"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
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
			StartDatabaseSnapshot(getDataStore(config.Store.Storage.StorageEngine))
			snapshotJobLastExecutedAt = time.Now()
		}

		time.Sleep(snapshotJobExecutionFrequency)
	}
}

func StartDatabaseSnapshot(store storage.DataStore) error {
	var recordCount int64 = 0
	var ssSizeInBytes int64 = 0
	var snapshotStartTime int64 = time.Now().UnixMilli()
	var err error

	snapshotservice := getSnapshotService(config.Store.Storage.StorageEngine)
	recordCount, ssSizeInBytes, err = snapshotservice.Snapshot(store)

	DatabaseInfoStats.Persistence.LastSnapshotTakenAt = utils.GetCurrentReadableTime()
	DatabaseInfoStats.Persistence.LastSnapshotLatency = fmt.Sprintf("%dms", time.Now().UnixMilli()-snapshotStartTime)
	DatabaseInfoStats.Persistence.SnapshotSizeInBytes = ssSizeInBytes
	DatabaseInfoStats.Keyspace.TotalKeyCount = recordCount

	return err
}

func RestoreDatabaseSnapshot(datastore storage.DataStore) {
	replayStartTime := time.Now().UnixMilli()

	snapshotservice := getSnapshotService(config.Store.Storage.StorageEngine)
	if ok, err := snapshotservice.ShouldRestore(); !ok {
		if err != nil {
			logger.Get().Info("Snapshot did not pass the restore check. Err=%v", err.Error())
		} else {
			logger.Get().Info("Snapshot restore is disabled. Skipping snapshot restore.")
		}
		return
	}

	keyCount, err := snapshotservice.Restore(datastore)

	replayLatency := fmt.Sprintf("%d ms", time.Now().UnixMilli()-replayStartTime)

	DatabaseInfoStats.Persistence.LastSnapshotReplayLatency = replayLatency
	DatabaseInfoStats.Persistence.TotalKeysReplayed = keyCount
	DatabaseInfoStats.Persistence.LastSnapshotReplayedAt = utils.GetCurrentReadableTime()

	if err != nil {
		logger.Get().Error("Snapshot restore failed, KeyOffset=%d, Err=%v", keyCount+1, err.Error())
		Shutdown(entity.ExitCodeStartupFailure)
	} else {
		logger.Get().Info("Snapshot restore done. Total %d keys replayed into DB", keyCount)
	}
}
