package lsm

import (
	"fmt"
	"sync"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage/lsm/memtable"
	"universum/storage/lsm/sstable"
	"universum/storage/lsm/wal"
)

const FlusherChanSize = 10
const WALRotaterChanSize = 10
const SSTableFlushRetryCount = 3

type LSMStore struct {
	memTable  memtable.MemTable
	sstables  []*sstable.SSTable
	mutex     sync.Mutex
	walWriter *wal.WALWriter
}

func CreateNewLSMStore(mtype string) *LSMStore {
	memtable := memtable.CreateNewMemTable(config.Store.Storage.LSM.MemtableStorageType)

	return &LSMStore{
		memTable: memtable,
		sstables: make([]*sstable.SSTable, 0),
	}
}

func (lsm *LSMStore) Initialize() error {
	sstableFiles, err := getAllSSTableFiles()
	if err != nil {
		return err
	}

	sstables := make([]*sstable.SSTable, len(sstableFiles))

	for i, filename := range sstableFiles {
		maxRecords := config.Store.Storage.LSM.BloomFilterMaxRecords
		fpRate := config.Store.Storage.LSM.BloomFalsePositiveRate

		sst, err := sstable.NewSSTable(filename, false, maxRecords, fpRate)
		if err != nil {
			return fmt.Errorf("failed to read SSTable %s:  %v", filename, err)
		}

		err = sst.LoadSSTableFromDisk()
		if err != nil {
			return fmt.Errorf("failed to load SSTable %s: %v", filename, err)
		}

		sstables[i] = sst
	}
	lsm.sstables = sstables

	lsm.walWriter, err = wal.NewWriter(config.Store.Storage.LSM.WriteAheadLogDirectory)
	if err != nil {
		return fmt.Errorf("failed to initialize write ahead logger: %v", err)
	}

	memtable.FlusherChan = make(chan memtable.MemTable, FlusherChanSize)
	memtable.WALRotaterChan = make(chan int64, WALRotaterChanSize)
	go lsm.MemtableBGFlusher() // start the background flusher job

	return nil
}

func (lsm *LSMStore) GetStoreType() string {
	return config.StorageEngineLSM
}

func (lsm *LSMStore) Exists(key string) (bool, uint32) {
	exists, code := lsm.memTable.Exists(key)
	if code == entity.CRC_RECORD_FOUND {
		return exists, code
	}

	// @TODO: Implement block cache to avoid reading from disk every time.

	for _, sst := range lsm.sstables {
		found, record, err := sst.FindRecord(key)
		if err != nil {
			return false, entity.CRC_DATA_READ_ERROR
		}

		if !found {
			continue // check in next sstable
		}

		if record.IsExpired() {
			return true, entity.CRC_RECORD_NOT_FOUND
		}
		return true, entity.CRC_RECORD_FOUND
	}

	return false, entity.CRC_RECORD_NOT_FOUND
}

func (lsm *LSMStore) Get(key string) (entity.Record, uint32) {
	record, code := lsm.memTable.Get(key)
	if code == entity.CRC_RECORD_FOUND {
		return record, code
	}

	// @TODO: Implement block cache to avoid reading from disk every time.

	for _, sst := range lsm.sstables {
		found, record, err := sst.FindRecord(key)
		if err != nil {
			return record, entity.CRC_DATA_READ_ERROR
		}

		if !found {
			continue // check in next sstable
		}

		if record.IsExpired() {
			return record, entity.CRC_RECORD_NOT_FOUND
		}
		return record, entity.CRC_RECORD_FOUND
	}

	return nil, entity.CRC_RECORD_NOT_FOUND
}

func (lsm *LSMStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	success, statusCode := lsm.memTable.Set(key, value, ttl)
	if !success && statusCode != entity.CRC_RECORD_UPDATED {
		return false, statusCode
	}

	err := lsm.walWriter.AddToWALBuffer(wal.OperationTypeSET, key, value, ttl)
	if err != nil {
		return false, entity.CRC_WAL_WRITE_FAILED
	}

	return true, entity.CRC_RECORD_UPDATED
}

func (lsm *LSMStore) Delete(key string) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) Append(key string, value string) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) MGet(keys []string) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) MDelete(keys []string) (map[string]interface{}, uint32) {
	// Implementation here
	return nil, 0
}

func (lsm *LSMStore) TTL(key string) (int64, uint32) {
	// Implementation here
	return 0, 0
}

func (lsm *LSMStore) Expire(key string, ttl int64) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) MemtableBGFlusher() error {
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Error("MemtableBGFlusher: Recovered from panic: %v", r)
			go lsm.MemtableBGFlusher() // Restart the flusher if it panics.
		}
	}()

	maxRecords := config.Store.Storage.LSM.BloomFilterMaxRecords
	fpRate := config.Store.Storage.LSM.BloomFalsePositiveRate

	for memtable := range memtable.FlusherChan {
		go func(container interface{}) {
			var sst *sstable.SSTable
			var err error

			newFileName := generateSSTableFileName()

			logger.Get().Info("BGFlusher: Flushing memtable to SSTable: %s", newFileName)

			for i := 0; i < SSTableFlushRetryCount; i++ {
				sst, err = sstable.NewSSTable(newFileName, true, maxRecords, fpRate)
				if err != nil {
					logger.Get().Error("[#%d] BGFlusher: failed to create new SSTable: %v", i+1, err)
					time.Sleep(10 * time.Millisecond)

					if i == SSTableFlushRetryCount-1 {
						// @TODO: handle error or consider shutting down the service if needed.
						logger.Get().Error("Background SSTable flush terminated after %d retries, Exiting", i+1)
						return
					}
					continue
				}
				break
			}

			for i := 0; i < SSTableFlushRetryCount; i++ {
				err = sst.FlushMemTableToSSTable(memtable)
				if err != nil {
					logger.Get().Error("[#%d] BGFlusher: failed to flush SSTable to disk: %v", i+1, err)
					time.Sleep(10 * time.Millisecond) // sleep for a while before retrying

					if i == SSTableFlushRetryCount-1 {
						// @TODO: handle error or consider shutting down the service if needed.
						logger.Get().Error("Background SSTable flush terminated after %d retries, Exiting", i+1)
						return
					}
					continue
				}
				break
			}

			lsm.mutex.Lock()
			lsm.sstables = append([]*sstable.SSTable{sst}, lsm.sstables...)
			lsm.mutex.Unlock()

		}(memtable)
	}
	return nil
}

func (lsm *LSMStore) Close() error {
	// @TODO handle more resource closures
	lsm.walWriter.Close()
	return nil
}
