package sstable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"universum/config"
	"universum/storage/lsm/memtable"
	"universum/utils"
)

func SetUpSSTableTests() {
	config.Store = config.GetSkeleton()
	config.Store.Logging.LogFileDirectory = "/tmp"
	config.Store.Storage.StorageEngine = config.StorageEngineLSM
	config.Store.Storage.MaxRecordSizeInBytes = 1048576
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeLB
	config.Store.Storage.LSM.MaxMemtableRecords = 1000
	config.Store.Storage.LSM.MaxMemtableDataSize = 1048576
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteBlockSize = 100
	config.Store.Storage.LSM.DataStorageDirectory = "/tmp"
	config.Store.Storage.LSM.BlockCompressionAlgo = config.CompressionAlgoLZ4
}

func TestNewSSTable(t *testing.T) {
	SetUpSSTableTests()

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, true, lsmCnf.MaxMemtableRecords, lsmCnf.BloomFalsePositiveRate)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath)

	if sst == nil {
		t.Fatalf("SSTable is nil")
	}

	if _, err := os.Stat(filepath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("SSTable file does not exist")
		}
	}

	if sst.fileptr == nil {
		t.Fatalf("SSTable file pointer is nil")
	}
}

func TestLoadSSTableFromDisk(t *testing.T) {
	SetUpSSTableTests()

	fileName := "test1.sst"
	cnf := config.Store.Storage.LSM

	sst, err := NewSSTable(fileName, true, 100, 0.01)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", cnf.DataStorageDirectory, fileName)))

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("value%d", i)
		mem.Set(key, val, int64(i))
	}

	err = sst.FlushMemTableToSSTable(mem)
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	// reset sst instance to populate again
	sst, _ = NewSSTable(fileName, false, 100, 0.01)

	err = sst.LoadSSTableFromDisk()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if (sst.RecordCount != 100) || (sst.RecordCount != sst.Metadata.NumRecords) {
		t.Fatalf("expected 100 records, got %d", sst.RecordCount)
	}

	if len(sst.Index) != 100 {
		t.Fatalf("expected 100 index entries, got %d", len(sst.Index))
	}

	if sst.BloomFilter.Size < 1 {
		t.Fatalf("expected non-zero bloom filter size, got %d", sst.BloomFilter.Size)
	}
}

func TestFlushMemTableToSSTable(t *testing.T) {
	SetUpSSTableTests()

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, true, lsmCnf.MaxMemtableRecords, lsmCnf.BloomFalsePositiveRate)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath)

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 10)
	mem.Set("key2", "value2", 10)

	err = sst.FlushMemTableToSSTable(mem)
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	if sst.RecordCount != 2 {
		t.Fatalf("Expected 2 records in SSTable, got %d", sst.RecordCount)
	}

	if sst.Metadata.DataSize != 603 {
		t.Fatalf("Expected 603B of size in SSTable metadata, got %dB", sst.Metadata.DataSize)
	}

	if sst.DataSize != 686 {
		t.Fatalf("Expected 686B of data in SSTable, got %dB", sst.DataSize)
	}

	_, err = os.Stat(filepath)
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("SSTable file does not exist at %s", filepath)
	}
}

func TestLoadBlock(t *testing.T) {
	SetUpSSTableTests()

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, true, lsmCnf.MaxMemtableRecords, lsmCnf.BloomFalsePositiveRate)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath)

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 10)
	mem.Set("key2", "value2", 10)

	_ = sst.FlushMemTableToSSTable(mem)

	blockOffset, blocksize := utils.UnpackNumbers(sst.Index["key1"])

	block, err := sst.LoadBlock(int64(blockOffset), int64(blocksize))
	block.PopulateRecordsInBlock()

	if err != nil {
		t.Fatalf("Failed to read block: %v", err)
	}

	if len(block.records) == 0 {
		t.Fatalf("Expected records in the block, found 0")
	}

	if len(block.data) == 0 {
		t.Fatalf("Expected data in the block, found empty")
	}

	key, value, err := block.ReadRecordAtOffset(block.index["key1"])

	if err != nil {
		t.Fatalf("Failed to read record from block: %v", err)
	}

	if string(key) != "key1" {
		t.Fatalf("Expected key1, got %s", key)
	}

	if value == nil {
		t.Fatalf("Expected value1, got %s", value)
	}
}
