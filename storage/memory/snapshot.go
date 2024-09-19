package memory

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
	"universum/storage"
	"universum/utils"
	"universum/utils/filesys"
)

var snapshotMutex sync.Mutex

type MemoryStoreSnapshotService struct{}

func (ms *MemoryStoreSnapshotService) StartDataBaseSnapshot(store storage.DataStore) (int64, int64, error) {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	shards := store.(*MemoryStore).GetAllShards()

	var masterRecordCount int64 = 0
	var shardRecordCount int64 = 0

	masterSnapshotFile := config.GetTransactionLogFilePath()
	tempSnapshotFile := fmt.Sprintf("%s/%s_qscRPQ6xHj", os.TempDir(), config.AppCodeName)

	tempFilePtr, _ := os.OpenFile(tempSnapshotFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	tempFilePtr.Close()

	translogBuff := storage.NewRecordTranslogBuffer()

	logger.Get().Info("Starting periodic record snapshot for all shards 1-%d", len(shards))

	for shardIndex := range shards {
		currentShard := shards[shardIndex]

		currentShard.GetData().Range(func(key interface{}, value interface{}) bool {
			record := value.(*entity.ScalarRecord)

			if record.Expiry <= utils.GetCurrentEPochTime() {
				return true // record already expired so skip
			}

			serialisedRecord, err := resp3.Encode(map[string]interface{}{
				"Key":    key,
				"Value":  record.Value,
				"LAT":    record.LAT,
				"Expiry": record.Expiry,
			})

			if err != nil {
				return true
			}

			done := translogBuff.AddToTranslogBuffer(serialisedRecord)
			if done {
				shardRecordCount++
			}

			return true
		})

		translogBuff.Flush(tempSnapshotFile)
		logger.Get().Debug("[Shard #%d] Periodic snapshot completed. %d records synced.", shardIndex+1, shardRecordCount)

		masterRecordCount += shardRecordCount
		shardRecordCount = 0
	}

	copyError := filesys.AtomicCopyFileContent(tempSnapshotFile, masterSnapshotFile)

	if copyError != nil {
		logger.Get().Error("Periodic DB snapshot creation failed for shard 1-%d; ERR=%v",
			len(shards), copyError.Error())

		return 0, 0, fmt.Errorf("snapshot failed: ERR=%v", copyError)
	}

	os.Remove(tempSnapshotFile)

	var snapshotSizeInBytes int64 = 0
	if fi, err := os.Stat(masterSnapshotFile); err == nil {
		snapshotSizeInBytes = fi.Size()
	}

	logger.Get().Info("Completed periodic DB snapshot for all shards 1-%d; Total Records=%d",
		len(shards), masterRecordCount)

	return masterRecordCount, snapshotSizeInBytes, nil
}

func (ms *MemoryStoreSnapshotService) ReplayDBRecordsFromSnapshot(datastore storage.DataStore) (int64, error) {
	var keycount int64 = 0

	buffer, err := ms.getRecordBuffer()
	if err != nil {
		return keycount, err
	}

	for {
		decoded, err := resp3.Decode(buffer)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				logger.Get().Error("failed to replay a commands into the datastore, "+
					"potentially errornous translog. Please fix to proceed: [%v]", err.Error())

				return keycount, err
			}
		}

		decodedMap, ok := decoded.(map[string]interface{})

		if !ok {
			logger.Get().Warn("failed to decode a command from RESP to map[string]interface, " +
				"potentially errornous translog record found.")

			continue
		}

		key, record := getRecordFromSerializedMap(decodedMap)
		scalarRecord, _ := record.(*entity.ScalarRecord)

		if scalarRecord.Expiry <= utils.GetCurrentEPochTime() {
			continue // skip as the record has already expired
		}

		success, _ := datastore.Set(key, scalarRecord.Value, scalarRecord.Expiry)

		if !success {
			logger.Get().Error("failed to replay a commands into the datastore, " +
				"potentially errornous translog or intermittent write failure, skipping.")

			continue
		}

		logger.Get().Debug("message replayed for [%5d] %-8s %#v", keycount, "SET", "ARGS")
		keycount++
	}

	return keycount, nil
}

func (ms *MemoryStoreSnapshotService) getRecordBuffer() (*bufio.Reader, error) {

	filepath := config.GetTransactionLogFilePath()
	filePtr, _ := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0777)
	defer filePtr.Close()

	snapshotSizeInBytes, err := filesys.GetFileSizeInBytes(filePtr)
	if err != nil {
		logger.Get().Error("failed to get the size of the snapshot file, ERR=%v", err.Error())
		return nil, err
	}

	buffer := bufio.NewReaderSize(filePtr, int(snapshotSizeInBytes))
	filePtr.Truncate(0)
	return buffer, nil
}
