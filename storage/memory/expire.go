package memory

import (
	"time"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage"
	"universum/utils"

	"golang.org/x/exp/rand"
)

const maxRecordDeletionLocalLimit int64 = 1000

func ExpireRandomSample(store storage.DataStore, shards [ShardCount]*Shard) int64 {
	var deletedCount int64 = 0

	randomGenerator := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	randomIndex := randomGenerator.Intn(len(shards))
	randomShard := shards[randomIndex]

	randomShard.GetData().Range(func(key interface{}, value interface{}) bool {
		record := value.(*entity.ScalarRecord)

		if record.Expiry < utils.GetCurrentEPochTime() {
			strkey, _ := key.(string)

			if deleted, _ := store.Delete(strkey); deleted {
				deletedCount++
			}
		}

		if deletedCount >= maxRecordDeletionLocalLimit {
			return false
		}

		return true
	})

	if deletedCount > 0 {
		logger.Get().Debug("RecordExpiryWorker:: ShardID=%d, Count=%d", randomIndex, deletedCount)
	}

	return deletedCount
}
