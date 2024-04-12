package engine

import (
	"log"
	"time"
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
	go snapshotWorker.startInMemoryDBSnapshot(snapshotChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range snapshotChan {

			log.Printf("Periodic record snapshot worker terminated. %v\n", failed.ExecutionErr)

			// Restart the worker if that has died
			go failed.startInMemoryDBSnapshot(snapshotChan)
			time.Sleep(10 * time.Second)
		}
	}()
}
