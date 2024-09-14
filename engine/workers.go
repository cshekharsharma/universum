package engine

import (
	"time"
	"universum/config"
	"universum/internal/logger"
)

func triggerPeriodicExpiryJob() {
	expworker := new(recordExpiryWorker)

	expiryChan := make(chan recordExpiryWorker, 1)
	go expworker.expireDeletedRecords(expiryChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range expiryChan {

			logger.Get().Warn("Periodic record expiry worker terminated. %v", failed.ExecutionErr)

			// Restart the worker if that has died
			go failed.expireDeletedRecords(expiryChan)
			time.Sleep(10 * time.Second)
		}
	}()
}

func triggerPeriodicSnapshotJob() {
	snapshotWorker := new(databaseSnapshotWorker)

	snapshotChan := make(chan databaseSnapshotWorker, 1)
	go snapshotWorker.startDatabaseSnapshot(snapshotChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range snapshotChan {

			logger.Get().Warn("Periodic record snapshot worker terminated. %v", failed.ExecutionErr)

			// Restart the worker if that has died
			go failed.startDatabaseSnapshot(snapshotChan)
			time.Sleep(10 * time.Second)
		}
	}()
}

func triggerPeriodicEvictionJob() {
	if config.GetRecordAutoEvictionPolicy() == EVICTION_POLICY_NONE {
		return
	}

	evictionWorker := new(autoEvictionWorker)

	evictionChan := make(chan autoEvictionWorker, 1)
	go evictionWorker.startAutoEviction(evictionChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range evictionChan {

			logger.Get().Warn("Periodic auto eviction worker terminated. %v", failed.ExecutionErr)

			// Restart the worker if that has died
			go failed.startAutoEviction(evictionChan)
			time.Sleep(10 * time.Second)
		}
	}()
}
