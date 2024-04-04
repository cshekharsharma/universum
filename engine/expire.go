package engine

import (
	"time"
	"universum/engine/entity"
	"universum/utils"
)

const MAX_RECORD_DELETION_LOCAL_LIMIT int64 = 1000
const MAX_RECORD_DELETION_FRACTION float64 = 0.2

var expiryJobLastExecutedAt time.Time
var expiryJobExecutionFrequency time.Duration = 3 * time.Second

func TriggerPeriodicExpiredRecordCleaup() {
	nextScheduledTime := expiryJobLastExecutedAt.Add(expiryJobExecutionFrequency)

	if nextScheduledTime.Compare(time.Now()) < 1 {
		expireDeletedRecords()
	}
}

func expireRandomSample() int64 {
	var deletedCount int64 = 0

	memStoreMutex.RLock()
	for pointer, record := range memoryStore {
		if isExpiredRecord(record) {
			if deleteByPointer(record, pointer) {
				deletedCount++
			}
		}

		if deletedCount >= MAX_RECORD_DELETION_LOCAL_LIMIT {
			break
		}
	}
	memStoreMutex.RUnlock()

	expiryJobLastExecutedAt = time.Now()
	return deletedCount
}

func isExpiredRecord(record *entity.Record) bool {
	expirationMutex.RLock()
	expiry, ok := expirationMap[record]
	expirationMutex.RUnlock()

	if !ok {
		return false
	}

	return expiry < utils.GetCurrentEPochTime()
}

func expireDeletedRecords() {
	var totalDeleted int64 = 0

	totalExpirable := int64(len(expirationMap))
	if totalExpirable == 0 {
		return
	}

	for {
		deleted := expireRandomSample()
		totalDeleted = deleted + totalDeleted
		ratio := float64(totalDeleted / totalExpirable)

		if ratio >= MAX_RECORD_DELETION_FRACTION {
			return
		}
	}

}
