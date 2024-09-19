package engine

import (
	"fmt"
	"time"
	"universum/config"
	"universum/storage"
	"universum/storage/memory"
)

var expiryJobLastExecutedAt time.Time
var expiryJobExecutionFrequency time.Duration

type recordExpiryWorker struct {
	ExecutionErr error
}

func (w *recordExpiryWorker) expireDeletedRecords(expiryChan chan<- recordExpiryWorker) (err error) {
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
		expiryChan <- *w
	}()

	store := getDataStore(config.GetStorageEngineType())

	var totalDeleted int64 = 0
	expiryJobLastExecutedAt = time.Now()

	for {
		nextScheduledTime := expiryJobLastExecutedAt.Add(expiryJobExecutionFrequency)

		if nextScheduledTime.Compare(time.Now()) < 1 {
			deleted := w.expireRandomSample(store)
			totalDeleted = deleted + totalDeleted

			expiryJobLastExecutedAt = time.Now()
			if deleted > 0 {
				continue
			}

		}

		time.Sleep(expiryJobExecutionFrequency)
	}
}

func (w *recordExpiryWorker) expireRandomSample(store storage.DataStore) int64 {
	switch store.GetStoreType() {
	case config.StorageTypeMemory:
		shards := store.(*memory.MemoryStore).GetAllShards()
		return memory.ExpireRandomSample(store.(*memory.MemoryStore), shards)

	default:
		return 0
	}
}
