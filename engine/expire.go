package engine

import (
	"universum/engine/entity"
	"universum/utils"
)

const MAX_RECORD_DELETION_LOCAL_LIMIT int64 = 1000
const MAX_RECORD_DELETION_FRACTION float64 = 20.0

func TriggerPeriodicExpiredRecordCleaup() {
	expireDeletedRecords()
}

func expireRandomSample() int64 {
	var deletedCount int64 = 0

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

	return deletedCount
}

func isExpiredRecord(record *entity.Record) bool {
	expiry, ok := expirationMap[record]
	if !ok {
		return false
	}

	return expiry < utils.GetCurrentEPochTime()
}

func expireDeletedRecords() {
	var totalDeleted int64 = 0

	totalExpirable := int64(len(expirationMap))
	if totalExpirable == 0 {
		totalExpirable = 1
	}

	for {
		deleted := expireRandomSample()
		totalDeleted = deleted + totalDeleted
		ratio := float64(totalDeleted / totalExpirable)

		if ratio < MAX_RECORD_DELETION_FRACTION {
			expireRandomSample()
		} else {
			break
		}
	}

}
