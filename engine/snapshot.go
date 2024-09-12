package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
	"universum/storage"
	"universum/utils"
	"universum/utils/filesys"
)

var snapshotMutex sync.Mutex

var snapshotJobLastExecutedAt time.Time
var snapshotJobExecutionFrequency time.Duration

type databaseSnapshotWorker struct {
	ExecutionErr error
}

func (w *databaseSnapshotWorker) startInMemoryDBSnapshot(snapshotChan chan<- databaseSnapshotWorker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				w.ExecutionErr = err
			} else {
				w.ExecutionErr = fmt.Errorf("panic happened with %v", r)
			}
		} else {
			w.ExecutionErr = err
		}

		// emit message to the worker channel that worker is dying.
		snapshotChan <- *w
	}()

	snapshotJobLastExecutedAt = time.Now()

	for {
		nextScheduledTime := snapshotJobLastExecutedAt.Add(expiryJobExecutionFrequency)

		if nextScheduledTime.Compare(time.Now()) < 1 {
			StartInMemoryDBSnapshot(GetMemstore())
			snapshotJobLastExecutedAt = time.Now()
		}

		time.Sleep(snapshotJobExecutionFrequency)
	}
}

func StartInMemoryDBSnapshot(memstore *storage.MemoryStore) {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	masterRecordCount := 0
	shardRecordCount := 0
	snapshotStartTime := time.Now().UnixMilli()

	masterSnapshotFile := config.GetTransactionLogFilePath()
	tempSnapshotFile := fmt.Sprintf("%s/%s_qscRPQ6xHj", os.TempDir(), config.APP_CODE_NAME)

	tempFilePtr, _ := os.OpenFile(tempSnapshotFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	tempFilePtr.Close()

	allShards := memstore.GetAllShards()
	translogBuff := newRecordTranslogBuffer()

	logger.Get().Info("Starting periodic record snapshot for all shards 1-%d", len(allShards))

	for shardIndex := range allShards {
		currentShard := allShards[shardIndex]

		currentShard.GetData().Range(func(key interface{}, value interface{}) bool {
			record := value.(*storage.ScalarRecord)

			serialisedRecord, err := resp3.Encode(map[string]interface{}{
				"Key":    key,
				"Value":  record.Value,
				"LAT":    record.LAT,
				"Expiry": record.Expiry,
			})

			if err != nil {
				return true
			}

			done := translogBuff.addToTranslogBuffer(serialisedRecord)
			if done {
				shardRecordCount++
			}

			return true
		})

		translogBuff.flush(tempSnapshotFile)
		logger.Get().Debug("[Shard #%d] Periodic snapshot completed. %d records synced.", shardIndex+1, shardRecordCount)

		masterRecordCount += shardRecordCount
		shardRecordCount = 0
	}

	copyError := filesys.AtomicCopyFileContent(tempSnapshotFile, masterSnapshotFile)

	if copyError != nil {
		logger.Get().Error("Periodic DB snapshot creation failed for shard 1-%d; ERR=%v",
			len(allShards), copyError.Error())

		return
	}

	os.Remove(tempSnapshotFile)

	DatabaseInfoStats.Persistence.LastSnapshotTakenAt = utils.GetCurrentReadableTime()
	DatabaseInfoStats.Persistence.LastSnapshotLatency = fmt.Sprintf("%dms", time.Now().UnixMilli()-snapshotStartTime)
	DatabaseInfoStats.Persistence.SnapshotSizeInBytes = 0

	if fi, err := os.Stat(masterSnapshotFile); err == nil {
		DatabaseInfoStats.Persistence.SnapshotSizeInBytes = fi.Size()
	}

	logger.Get().Info("Completed periodic DB snapshot for all shards 1-%d; Total Records=%d",
		len(allShards), masterRecordCount)
}

func ReplayDBRecordsFromSnapshot() (int64, error) {
	var keycount int64 = 0
	var replayStartTime int64 = time.Now().UnixMilli()

	filepath := config.GetTransactionLogFilePath()
	filePtr, _ := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0777)
	defer filePtr.Close()

	snapshotSizeInBytes, err := filesys.GetFileSizeInBytes(filePtr)
	if err != nil {
		logger.Get().Error("failed to get the size of the snapshot file, ERR=%v", err.Error())
		return keycount, err
	}

	buffer := bufio.NewReaderSize(filePtr, int(snapshotSizeInBytes))

	for {
		decoded, err := resp3.Decode(buffer)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				logger.Get().Debug("failed to replay a commands into the memorystore, "+
					"potentially errornous translog. Please fix to proceed: [%v]", err.Error())

				return keycount, err
			}
		}

		decodedMap, ok := decoded.(map[string]interface{})

		if !ok {
			logger.Get().Debug("failed to decode a command from RESP to map[string]interface, " +
				"potentially errornous translog record found.")

			continue
		}

		key, record := getRecordFromSerializedMap(decodedMap)
		scalarRecord, _ := record.(*storage.ScalarRecord)

		if scalarRecord.Expiry <= utils.GetCurrentEPochTime() {
			continue // skip as the record has already expired
		}

		command := &entity.Command{
			Name: COMMAND_SET,
			Args: []interface{}{
				key,
				scalarRecord.Value,
				scalarRecord.Expiry - utils.GetCurrentEPochTime(),
			},
		}

		_, execErr := executeCommand(context.Background(), command)

		if execErr != nil {
			logger.Get().Debug("failed to replay a commands into the memorystore, " +
				"potentially errornous translog or intermittent write failure.")

			continue
		}

		logger.Get().Debug("message replayed for [%5d] %-8s %#v", keycount, command.Name, command.Args)
		keycount++
	}

	filePtr.Truncate(0) // truncate the translog for fresh records
	replayLatency := fmt.Sprintf("%d ms", time.Now().UnixMilli()-replayStartTime)

	DatabaseInfoStats.Persistence.LastSnapshotReplayLatency = replayLatency
	DatabaseInfoStats.Persistence.TotalKeysReplayed = keycount
	DatabaseInfoStats.Persistence.LastSnapshotReplayedAt = utils.GetCurrentReadableTime()

	return keycount, nil
}

func getRecordFromSerializedMap(recordMap map[string]interface{}) (string, storage.Record) {
	record := new(storage.ScalarRecord)
	var key string

	if k, ok := recordMap["Key"]; ok {
		key, _ = k.(string)
	}

	if val, ok := recordMap["Value"]; ok {
		record.Value = val
		record.Type = storage.GetTypeEncoding(val)
	}

	if lat, ok := recordMap["LAT"]; ok {
		record.LAT, _ = lat.(int64)
	}

	if expiry, ok := recordMap["Expiry"]; ok {
		record.Expiry, _ = expiry.(int64)
	}

	return key, record
}
