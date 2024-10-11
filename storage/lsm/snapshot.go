package lsm

import (
	"fmt"
	"sync"
	"universum/config"
	"universum/storage"
	"universum/storage/lsm/wal"
)

var (
	restoreMutex sync.Mutex
)

type LSMStoreSnapshotService struct{}

func (ms *LSMStoreSnapshotService) Snapshot(datastore storage.DataStore) (int64, int64, error) {
	return 0, 0, nil
}

func (ms *LSMStoreSnapshotService) Restore(datastore storage.DataStore) (int64, error) {
	restoreMutex.Lock()
	defer restoreMutex.Unlock()

	walReader, err := wal.NewReader(config.Store.Storage.LSM.WriteAheadLogDirectory)
	if err != nil {
		return 0, fmt.Errorf("failed to create WAL reader: %v", err)
	}

	keycount, err := walReader.RestoreFromWAL(datastore.(*LSMStore).memTable)
	if err != nil {
		return keycount, fmt.Errorf("failed to restore from WAL: %v", err)
	}

	datastore.(*LSMStore).memTable.Truncate()
	return keycount, nil
}

func (ms *LSMStoreSnapshotService) ShouldRestore() (bool, error) {
	return true, nil
}
