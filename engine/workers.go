package engine

import (
	"log"
	"time"
	"universum/config"
)

func triggerPeriodicExpiryJob() {
	expworker := new(recordExpiryWorker)

	expiryChan := make(chan recordExpiryWorker, 1)
	go expworker.expireDeletedRecords(expiryChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range expiryChan {

			log.Printf("Periodic record expiry worker terminated. %v\n", failed.ExecutionErr)

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

			log.Printf("Periodic record snapshot worker terminated. %v\n", failed.ExecutionErr)

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

			log.Printf("Periodic auto eviction worker terminated. %v\n", failed.ExecutionErr)

			// Restart the worker if that has died
			go failed.startAutoEviction(evictionChan)
			time.Sleep(10 * time.Second)
		}
	}()
}
