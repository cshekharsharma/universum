package memory

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	"universum/compression"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
	"universum/storage"
	"universum/utils"
	"universum/utils/filesys"
)

var (
	snapshotMutex        sync.Mutex
	errSnapshotFileEmpty error = errors.New("snapshot file is empty")
)

const (
	SnapshotFileName   string = "snapshot.db"
	MaxBlockBufferSize int    = 4 * 1024 * 1024 // 4MB
)

type MemoryStoreSnapshotService struct{}

func (ms *MemoryStoreSnapshotService) StartDatabaseSnapshot(store storage.DataStore) (int64, int64, error) {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	snapshotStartedAt := time.Now().UnixMilli()
	shards := store.(*MemoryStore).GetAllShards()

	var masterRecordCount int64 = 0
	var shardRecordCount int64 = 0

	masterSnapshotFilePath := GetSnapshotFilePath()

	randomSuffix := utils.GetRandomString(10)
	tempSnapshotFilePath := fmt.Sprintf("%s/%s_%s_snapshot", os.TempDir(), config.AppCodeName, randomSuffix)

	tempFilePtr, err := os.OpenFile(tempSnapshotFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temporary snapshot file: %v", err)
	}
	defer tempFilePtr.Close()

	var buffer bytes.Buffer
	compressor := compression.GetCompressor(&compression.Options{
		Writer:          tempFilePtr,
		AutoCloseWriter: false,
	})

	logger.Get().Info("Starting periodic record snapshot for all shards 1-%d", len(shards))

	for shardIndex := range shards {
		currentShard := shards[shardIndex]

		currentShard.GetData().Range(func(key interface{}, value interface{}) bool {
			record := value.(*entity.ScalarRecord)

			if record.Expiry < utils.GetCurrentEPochTime() {
				return true // record already expired so skip
			}

			serialisedRecord, err := resp3.Encode(map[string]interface{}{
				"Key":    key,
				"Value":  record.Value,
				"LAT":    record.LAT,
				"Expiry": record.Expiry,
			})

			if err != nil {
				logger.Get().Warn("Failed to serialise record for snapshot (skipping); ERR=%v", err)
				return true
			}

			_, err = buffer.Write([]byte(serialisedRecord))
			if err != nil {
				return true
			}
			shardRecordCount++

			if buffer.Len() >= MaxBlockBufferSize {
				if err := compressor.CompressAndWrite(buffer.Bytes()); err != nil {
					logger.Get().Error("Failed to write compressed block during snapshot: %v", err)
					return false // Stop on error
				}
				buffer.Reset()
			}

			return true
		})

		logger.Get().Debug("[Shard #%d] Periodic snapshot completed. %d records synced.", shardIndex+1, shardRecordCount)

		masterRecordCount += shardRecordCount
		shardRecordCount = 0
	}

	compressor.Close()
	copyError := filesys.AtomicCopyFileContent(tempSnapshotFilePath, masterSnapshotFilePath)

	if copyError != nil {
		logger.Get().Error("Periodic DB snapshot creation failed for shard 1-%d; ERR=%v",
			len(shards), copyError.Error())

		return 0, 0, fmt.Errorf("snapshot failed: ERR=%v", copyError)
	}

	os.Remove(tempSnapshotFilePath)

	var snapshotSizeInBytes int64 = 0
	if fi, err := os.Stat(masterSnapshotFilePath); err == nil {
		snapshotSizeInBytes = fi.Size()
	}

	snapshotEndedAt := time.Now().UnixMilli()
	logger.Get().Info("Periodic DB snapshot completed for all shards 1-%d; Total Records=%d; TimeTaken: %dms",
		len(shards), masterRecordCount, (snapshotEndedAt - snapshotStartedAt))

	return masterRecordCount, snapshotSizeInBytes, nil
}

func (ms *MemoryStoreSnapshotService) ReplayDBRecordsFromSnapshot(datastore storage.DataStore) (int64, error) {
	var keycount int64 = 0

	buffer, fileptr, err := ms.getRecordBuffer()
	if err != nil {
		fileptr.Close()

		if errors.Is(err, errSnapshotFileEmpty) {
			return keycount, nil
		} else {
			return keycount, err
		}
	}

	defer fileptr.Truncate(0)
	defer fileptr.Close()

	if compression.IsCompressionEnabled() {
		c := compression.GetCompressor(&compression.Options{})
		decodedStream, err := c.DecompressAndRead(int64(MaxBlockBufferSize))

		if err != nil {
			logger.Get().Error("failed to decompress the snapshot file, ERR=%v", err.Error())
			return keycount, err
		}

		buffer = bufio.NewReader(bytes.NewReader(decodedStream))
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

		ttl := scalarRecord.Expiry - utils.GetCurrentEPochTime()
		success, _ := datastore.Set(key, scalarRecord.Value, ttl)

		if !success {
			logger.Get().Error("failed to replay a commands into the datastore, " +
				"potentially errornous translog or intermittent write failure, skipping.")

			continue
		}

		logger.Get().Debug("message replayed for [%5d] SET %s %v %v", keycount, key, scalarRecord.Value, ttl)
		keycount++
	}

	return keycount, nil
}

func (ms *MemoryStoreSnapshotService) getRecordBuffer() (*bufio.Reader, *os.File, error) {

	filepath := GetSnapshotFilePath()
	filePtr, _ := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0777)

	snapshotSizeInBytes, err := filesys.GetFileSizeInBytes(filePtr)
	if err != nil {
		logger.Get().Error("failed to get the size of the snapshot file, ERR=%v", err.Error())
		return nil, filePtr, err
	}

	if snapshotSizeInBytes == 0 {
		return nil, filePtr, errSnapshotFileEmpty
	}

	buffer := bufio.NewReaderSize(filePtr, int(snapshotSizeInBytes))
	return buffer, filePtr, nil
}

func GetSnapshotFilePath() string {
	masterSnapshotFile := fmt.Sprintf("%s/%s", config.Store.Storage.Memory.SnapshotFileDirectory, SnapshotFileName)
	return filepath.Clean(masterSnapshotFile)
}
