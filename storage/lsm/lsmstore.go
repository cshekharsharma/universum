package lsm

import (
	"fmt"
	"sync"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage/lsm/compaction"
	"universum/storage/lsm/memtable"
	"universum/storage/lsm/sstable"
	"universum/storage/lsm/wal"
	"universum/utils"
)

const FlusherChanSize = 10
const WALRotaterChanSize = 10
const SSTableFlushRetryCount = 3

type LSMStore struct {
	memTable  memtable.MemTable
	sstables  []*sstable.SSTable
	walWriter *wal.WALWriter
	compactor *compaction.Compactor
	mutex     sync.Mutex
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

	lsm.compactor = compaction.NewCompactor()
	go lsm.compactor.Compact() // start the background compaction job

	sstable.BlockCacheStore = sstable.NewBlockCache()
	return nil
}

func (lsm *LSMStore) GetStoreType() string {
	return config.StorageEngineLSM
}

func (lsm *LSMStore) Exists(key string) (bool, uint32) {
	exists, code := lsm.memTable.Exists(key)
	if yes, _ := utils.ExistsInList(code, []uint32{entity.CRC_RECORD_EXPIRED, entity.CRC_RECORD_TOMBSTONED}); yes {
		return false, entity.CRC_RECORD_NOT_FOUND
	}

	if code == entity.CRC_RECORD_FOUND {
		return exists, code
	}

	for _, sst := range lsm.sstables {
		found, record, err := sst.FindRecord(key)
		if err != nil {
			return false, entity.CRC_DATA_READ_ERROR
		}

		if !found || record == nil {
			continue // check in next sstable
		}

		if record.IsTombstoned() || record.IsExpired() {
			return false, entity.CRC_RECORD_NOT_FOUND
		}
		return true, entity.CRC_RECORD_FOUND
	}

	return false, entity.CRC_RECORD_NOT_FOUND
}

func (lsm *LSMStore) Get(key string) (entity.Record, uint32) {
	record, code := lsm.memTable.Get(key)
	if yes, _ := utils.ExistsInList(code, []uint32{entity.CRC_RECORD_EXPIRED, entity.CRC_RECORD_TOMBSTONED}); yes {
		return nil, entity.CRC_RECORD_NOT_FOUND
	}

	if code == entity.CRC_RECORD_FOUND {
		return record, code
	}

	// @TODO: Implement block cache to avoid reading from disk every time.

	for _, sst := range lsm.sstables {
		found, record, err := sst.FindRecord(key)
		if err != nil {
			return record, entity.CRC_DATA_READ_ERROR
		}

		if !found || record == nil {
			continue // check in next sstable
		}

		if record.IsTombstoned() || record.IsExpired() {
			return record, entity.CRC_RECORD_NOT_FOUND
		}
		return record, entity.CRC_RECORD_FOUND
	}

	return nil, entity.CRC_RECORD_NOT_FOUND
}

func (lsm *LSMStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	success, statusCode := lsm.memTable.Set(key, value, ttl, entity.RecordStateActive)
	if !success && statusCode != entity.CRC_RECORD_UPDATED {
		return false, statusCode
	}

	err := lsm.walWriter.AddToWALBuffer(key, value, ttl, entity.RecordStateActive)
	if err != nil {
		return false, entity.CRC_WAL_WRITE_FAILED
	}

	return true, entity.CRC_RECORD_UPDATED
}

func (lsm *LSMStore) Delete(key string) (bool, uint32) {
	lsm.memTable.Delete(key)

	err := lsm.walWriter.AddToWALBuffer(key, 0, time.Now().Unix(), entity.RecordStateTombstoned)
	if err != nil {
		return false, entity.CRC_WAL_WRITE_FAILED
	}

	return true, entity.CRC_RECORD_DELETED
}

func (lsm *LSMStore) IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32) {
	val, code := lsm.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, code
	}

	record := val.(*entity.ScalarRecord)
	if !utils.IsInteger(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	var newValue int64
	oldValue := record.Value.(int64)

	if isIncr {
		newValue = int64(oldValue) + offset
	} else {
		newValue = int64(oldValue) - offset
	}

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	didSet, setcode := lsm.Set(key, newValue, ttl)

	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return newValue, entity.CRC_RECORD_UPDATED
}

func (lsm *LSMStore) Append(key string, value string) (int64, uint32) {
	val, code := lsm.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return config.InvalidNumericValue, code
	}

	record := val.(*entity.ScalarRecord)
	if !utils.IsString(record.Value) {
		return config.InvalidNumericValue, entity.CRC_INCR_INVALID_TYPE
	}

	newValue := record.Value.(string) + value
	ttl := record.Expiry - utils.GetCurrentEPochTime()

	didSet, setcode := lsm.Set(key, newValue, ttl)
	if !didSet {
		return config.InvalidNumericValue, setcode
	}

	return int64(len(newValue)), entity.CRC_RECORD_UPDATED
}

func (lsm *LSMStore) MGet(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		record, code := lsm.Get(keys[idx])

		if _, ok := record.(*entity.ScalarRecord); ok {
			responseMap[keys[idx]] = map[string]interface{}{
				"Value": record.(*entity.ScalarRecord).Value,
				"Code":  code,
			}
		} else {
			responseMap[keys[idx]] = map[string]interface{}{
				"Value": nil,
				"Code":  code,
			}
		}
	}

	return responseMap, entity.CRC_MGET_COMPLETED
}

func (lsm *LSMStore) MSet(kvMap map[string]interface{}) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for key, value := range kvMap {
		didSet, code := lsm.Set(key, value, 0)
		if code == entity.CRC_RECORD_TOO_BIG || code == entity.CRC_INVALID_DATATYPE {
			fmt.Printf("Key %s failed due to %d\n", key, code)
		}
		responseMap[key] = didSet
	}

	return responseMap, entity.CRC_MSET_COMPLETED
}

func (lsm *LSMStore) MDelete(keys []string) (map[string]interface{}, uint32) {
	responseMap := make(map[string]interface{})

	for idx := range keys {
		deleted, _ := lsm.Delete(keys[idx])
		responseMap[keys[idx]] = deleted
	}

	return responseMap, entity.CRC_MDEL_COMPLETED
}

func (lsm *LSMStore) TTL(key string) (int64, uint32) {
	val, code := lsm.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return 0, code
	}

	record := val.(*entity.ScalarRecord)

	ttl := record.Expiry - utils.GetCurrentEPochTime()
	return ttl, entity.CRC_RECORD_FOUND
}

func (lsm *LSMStore) Expire(key string, ttl int64) (bool, uint32) {
	val, code := lsm.Get(key)

	if code != entity.CRC_RECORD_FOUND {
		return false, code
	}

	record := val.(*entity.ScalarRecord)
	return lsm.Set(key, record.Value, ttl)
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

			lsm.mutex.Lock()
			defer lsm.mutex.Unlock()

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

			lsm.sstables = append([]*sstable.SSTable{sst}, lsm.sstables...)
			lsm.compactor.CompactLevel(sst.Metadata.CompactionLevel)
		}(memtable)
	}
	return nil
}

func (lsm *LSMStore) Close() error {
	// @TODO handle more resource closures
	lsm.walWriter.Close()
	return nil
}
