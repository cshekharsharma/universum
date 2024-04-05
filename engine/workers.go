package engine

import (
	"log"
	"time"
)

func triggerPeriodicExpiryJob() {
	expworker := new(periodicRecordExpiryWorker)

	expiryChan := make(chan periodicRecordExpiryWorker, 1)
	go expworker.expireDeletedRecords(expiryChan)

	// Go routine to respawn the worker
	go func() {
		for failed := range expiryChan {

			log.Printf("Periodic record expiry worker terminated. %v\n", failed.ExecutionErr)

			// Restart the worker overseer if that has died
			go failed.expireDeletedRecords(expiryChan)
			time.Sleep(10 * time.Second)
		}

	}()
}
