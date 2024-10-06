package lsm

import (
	"fmt"
	"universum/config"
	"universum/storage"
	"universum/storage/lsm/wal"
)

type LSMStoreSnapshotService struct{}

func (ms *LSMStoreSnapshotService) Snapshot(store storage.DataStore) (int64, int64, error) {
	return 0, 0, nil
}

func (ms *LSMStoreSnapshotService) Restore(datastore storage.DataStore) (int64, error) {
	walReader, err := wal.NewReader(config.Store.Storage.LSM.WriteAheadLogDirectory)
	if err != nil {
		return 0, fmt.Errorf("failed to create WAL reader: %v", err)
	}

	return walReader.RestoreFromWAL(datastore.(*LSMStore).memTable)
}

func (ms *LSMStoreSnapshotService) ShouldRestore() (bool, error) {
	return true, nil
}
