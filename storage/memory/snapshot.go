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
	MaxBlockBufferSize int64  = 4 * 1024 * 1024 // 4MB
)

type MemoryStoreSnapshotService struct{}

func (ms *MemoryStoreSnapshotService) Snapshot(store storage.DataStore) (int64, int64, error) {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	snapshotStartedAt := time.Now().UnixMilli()
	shards := store.(*MemoryStore).GetAllShards()

	var masterRecordCount int64 = 0
	var shardRecordCount int64 = 0

	masterSnapshotFilePath := getSnapshotFilePath()

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

			if buffer.Len() >= int(MaxBlockBufferSize) {
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

	if buffer.Len() > 0 {
		if err := compressor.CompressAndWrite(buffer.Bytes()); err != nil {
			logger.Get().Error("Failed to write remaining compressed block during snapshot: %v", err)
			return 0, 0, err
		}
		buffer.Reset()
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

func (ms *MemoryStoreSnapshotService) Restore(datastore storage.DataStore) (int64, error) {
	var keycount int64 = 0
	snapshotFilePath := getSnapshotFilePath()

	filePtr, err := ms.getSnapshotFilePtr(snapshotFilePath)
	if err != nil {
		filePtr.Close()

		if errors.Is(err, errSnapshotFileEmpty) {
			return keycount, nil
		} else {
			return keycount, err
		}
	}

	defer filePtr.Close()

	c := compression.GetCompressor(&compression.Options{
		Reader: filePtr,
	})

	var partialBuffer []byte     // Holds unprocessed raw bytes for the next chunk
	var backupChunkBuffer []byte // Duplicate buffer to track raw bytes

	for {
		chunkBuffer, err := c.DecompressAndRead(int64(MaxBlockBufferSize))
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			logger.Get().Error("Failed to decompress the snapshot file, ERR=%v", err.Error())
			return keycount, err
		}

		chunkBuffer = append(partialBuffer, chunkBuffer...)
		partialBuffer = nil

		backupChunkBuffer = make([]byte, len(chunkBuffer))
		copy(backupChunkBuffer, chunkBuffer)
		lastSuccessfulOffset := 0

		buffer := bufio.NewReaderSize(bytes.NewReader(chunkBuffer), len(chunkBuffer))

		for {
			decoded, err := resp3.Decode(buffer)
			currentOffset := len(backupChunkBuffer) - buffer.Buffered()

			if err != nil {
				if (err == io.EOF) || (err == io.ErrUnexpectedEOF) || (err == io.ErrShortBuffer) {
					logger.Get().Warn("Potentially incomplete message read. Will retry with autofix [%v]", err.Error())

					// Calculating the remaining bytes based on the last successfully processed message
					// and storing remaining bytes to prepend with the next chunk in the next iteration
					remainingBytes := backupChunkBuffer[lastSuccessfulOffset:]
					partialBuffer = append([]byte{}, remainingBytes...)

					break
				} else {
					logger.Get().Error("Failed to replay a commands into the datastore, "+
						"potentially erroneous record. [ERR=%v]", err.Error())

					return keycount, err
				}
			}

			if decodedMap, ok := decoded.(map[string]interface{}); ok {
				if len(decodedMap) < 4 {
					remainingBytes := backupChunkBuffer[lastSuccessfulOffset:]
					partialBuffer = append([]byte{}, remainingBytes...)
					break
				}

				lastSuccessfulOffset = currentOffset // updated last successful offset
				key, record := getRecordFromSerializedMap(decodedMap)
				scalarRecord, _ := record.(*entity.ScalarRecord)

				if scalarRecord.Expiry <= utils.GetCurrentEPochTime() {
					continue
				}

				ttl := scalarRecord.Expiry - utils.GetCurrentEPochTime()
				success, statuscode := datastore.Set(key, scalarRecord.Value, ttl)

				if success {
					logger.Get().Debug("Replayed [%5d] SET %s %v %v", keycount+1, key, scalarRecord.Value, ttl)
					keycount++
				} else {
					logger.Get().Warn("Failed to restore the record [code=%d], skipping.", statuscode)
				}
			} else {
				logger.Get().Warn("Failed to decode a command from RESP to map[string]interface{}, " +
					"potentially erroneous translog record found.")
			}
		}
	}

	return keycount, nil
}

func (ms *MemoryStoreSnapshotService) ShouldRestore() (bool, error) {
	return config.Store.Storage.Memory.RestoreSnapshotOnStart, nil
}

func (ms *MemoryStoreSnapshotService) getSnapshotFilePtr(filepath string) (*os.File, error) {
	filePtr, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		logger.Get().Error("failed to open the snapshot file, ERR=%v", err.Error())
		return nil, err
	}

	snapshotSizeInBytes, err := filesys.GetFileSizeInBytes(filePtr)
	if err != nil {
		logger.Get().Error("failed to get the size of the snapshot file, ERR=%v", err.Error())
		return filePtr, err
	}

	if snapshotSizeInBytes == 0 {
		return filePtr, errSnapshotFileEmpty
	}

	return filePtr, nil
}

func getSnapshotFilePath() string {
	directory := config.Store.Storage.Memory.SnapshotFileDirectory
	masterSnapshotFile := fmt.Sprintf("%s/%s", directory, SnapshotFileName)
	return filepath.Clean(masterSnapshotFile)
}
