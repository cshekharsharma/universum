package engine

import (
	"fmt"
	"math/rand"
	"time"
	"universum/internal/logger"
	"universum/storage"
)

const shardcount uint32 = storage.ShardCount
const maxRecordDeletionLocalLimit int64 = 1000

var expiryJobLastExecutedAt time.Time
var expiryJobExecutionFrequency time.Duration = 2 * time.Second

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

	memorystore := GetMemstore()
	shards := memorystore.GetAllShards()

	var totalDeleted int64 = 0
	expiryJobLastExecutedAt = time.Now()

	for {
		nextScheduledTime := expiryJobLastExecutedAt.Add(expiryJobExecutionFrequency)

		if nextScheduledTime.Compare(time.Now()) < 1 {
			deleted := w.expireRandomSample(memorystore, shards)
			totalDeleted = deleted + totalDeleted

			if deleted > 0 {
				continue
			}
		}

		time.Sleep(expiryJobExecutionFrequency)
	}
}

func (w *recordExpiryWorker) expireRandomSample(memstore *storage.MemoryStore, shards [shardcount]*storage.Shard) int64 {
	var deletedCount int64 = 0

	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := randomGenerator.Intn(len(shards))
	randomShard := shards[randomIndex]

	randomShard.GetData().Range(func(key interface{}, value interface{}) bool {
		record := value.(*storage.ScalarRecord)

		if record.Expiry < time.Now().Unix() {
			strkey, _ := key.(string)

			if deleted, _ := memstore.Delete(strkey); deleted {
				deletedCount++
			}
		}

		if deletedCount >= maxRecordDeletionLocalLimit {
			return false
		}

		return true
	})

	expiryJobLastExecutedAt = time.Now()

	if deletedCount > 0 {
		logger.Get().Debug("Few keys deleted. ShardID=%d, Count=%d", randomIndex, deletedCount)
	}

	return deletedCount
}
