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
)

const FlusherChanSize = 10
const SSTableFlushRetryCount = 3

type LSMStore struct {
	memTable memtable.MemTable
	sstables []*sstable.SSTable // or another suitable data structure
	mutex    sync.Mutex
	wal      *WriteAheadLogger
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
		sst, err := sstable.NewSSTable(filename, false, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to load SSTable file %s: %v", filename, err)
		}

		err = sst.LoadMetadata()
		if err != nil {
			return fmt.Errorf("failed to load metadata for SSTable %s: %v", filename, err)
		}

		err = sst.LoadBloomFilter()
		if err != nil {
			return fmt.Errorf("failed to load Bloom filter for SSTable %s: %v", filename, err)
		}

		err = sst.LoadIndex()
		if err != nil {
			return fmt.Errorf("failed to load index for SSTable %s: %v", filename, err)
		}

		sstables[i] = sst
	}

	lsm.wal, err = NewWAL(config.Store.Storage.LSM.WriteAheadLogDirectory)
	if err != nil {
		return fmt.Errorf("failed to initialize write ahead logger: %v", err)
	}

	memtable.FlusherChan = make(chan memtable.MemTable, FlusherChanSize)
	go lsm.MemtableBGFlusher() // start the background flusher job

	return nil
}

func (lsm *LSMStore) GetStoreType() string {
	return config.StorageTypeLSM
}

func (lsm *LSMStore) Exists(key string) (bool, uint32) {
	// Implementation here
	return false, 0
}

func (lsm *LSMStore) Get(key string) (entity.Record, uint32) {
	record, code := lsm.memTable.Get(key)
	if code == entity.CRC_RECORD_FOUND {
		return record, code
	}

	// @TODO: Implement block cache
	// val, found = lsm.blockCache.Get(key)
	// if found {
	// 	return val, nil
	// }

	for _, sst := range lsm.sstables {
		if sst.BloomFilter != nil && !sst.BloomFilter.Exists(key) {
			continue // move to next SST if bloom filter says no
		}

		blockOffset, ok := sst.Index[key]
		if ok && blockOffset != 0 {
			block, err := sst.ReadBlock(blockOffset)
			if err != nil {
				return nil, entity.CRC_DATA_READ_ERROR
			}

			record, err := block.GetRecord(key)
			if err != nil {
				return nil, entity.CRC_RECORD_NOT_FOUND
			}

			return record, entity.CRC_RECORD_FOUND
		}
	}

	return nil, entity.CRC_RECORD_NOT_FOUND
}

func (lsm *LSMStore) Set(key string, value interface{}, ttl int64) (bool, uint32) {
	success, statusCode := lsm.memTable.Set(key, value, ttl)
	if !success && statusCode != entity.CRC_RECORD_UPDATED {
		return false, statusCode
	}

	err := lsm.wal.AddToWALBuffer(key, value, ttl)
	if err != nil {
		return false, entity.CRC_AOF_WRITE_FAILED
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

	for memtable := range memtable.FlusherChan {
		go func(container interface{}) {
			var sst *sstable.SSTable
			var err error

			newFileName := generateSSTableFileName()

			for i := 0; i < SSTableFlushRetryCount; i++ {
				sst, err = sstable.NewSSTable(newFileName, true, 0, 0)
				if err != nil {
					logger.Get().Error("[#%d] BGFlusher: Failed to create new SSTable: %v", i, err)
					time.Sleep(10 * time.Millisecond)

					if i == SSTableFlushRetryCount-1 {
						// @TODO: handle error or consider shutting down the service if needed.
						logger.Get().Error("BGFlusher: Failed to create new SSTable after %d retries: %v", i-1, err)
						return
					}
					continue
				}
				break
			}

			for i := 0; i < SSTableFlushRetryCount; i++ {
				err = sst.FlushMemTableToSSTable(memtable)
				if err != nil {
					logger.Get().Error("[#%d] BGFlusher: Failed to flush SSTable to disk: %v", i, err)
					time.Sleep(10 * time.Millisecond)

					if i == SSTableFlushRetryCount-1 {
						// @TODO: handle error or consider shutting down the service if needed.
						logger.Get().Error("BGFlusher: Failed to create new SSTable after %d retries: %v", i-1, err)
						return
					}
					continue
				}
				break
			}

			lsm.mutex.Lock()
			lsm.sstables = append(lsm.sstables, sst)
			lsm.mutex.Unlock()

		}(memtable)
	}
	return nil
}
