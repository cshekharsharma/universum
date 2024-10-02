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
	config.Store.Storage.StorageEngine = config.StorageEngineLSM
	config.Store.Storage.MaxRecordSizeInBytes = 1048576
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeLB
	config.Store.Storage.LSM.MaxMemtableRecords = 1000
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteBlockSize = 100
	config.Store.Storage.LSM.DataStorageDirectory = "/tmp"
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

	if sst.Metadata.DataSize != 621 {
		t.Fatalf("Expected 621B of size in SSTable metadata, got %dB", sst.Metadata.DataSize)
	}

	if sst.DataSize != 705 {
		t.Fatalf("Expected 705B of data in SSTable, got %dB", sst.DataSize)
	}

	f, _ := os.Stat(filepath)
	if f.Size() != 705 {
		t.Fatalf("Expected 705B of data in SSTable file, got %dB", f.Size())
	}
}

func TestReadBlock(t *testing.T) {
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

	block, err := sst.ReadBlock(int64(blockOffset), int64(blocksize))
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

func TestLoadMetadata(t *testing.T) {
	sst, err := NewSSTable("test.sst", true, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove("test.sst")

	mem := memtable.CreateNewMemTable("skiplist").(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 0)

	err = sst.FlushMemTableToSSTable(mem)
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	err = sst.WriteMetadata()
	if err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	sst.fileptr.Close()

	// Reopen the file in read mode to test loading metadata
	sst, err = NewSSTable("test.sst", false, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to open SSTable: %v", err)
	}

	err = sst.LoadMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if sst.Metadata.NumRecords != 1 {
		t.Fatalf("Expected metadata to reflect 1 record, got %d", sst.Metadata.NumRecords)
	}
}

func TestLoadIndex(t *testing.T) {
	sst, err := NewSSTable("test.sst", true, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove("test.sst")

	mem := memtable.CreateNewMemTable("skiplist").(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 0)
	mem.Set("key2", "value2", 0)

	err = sst.FlushMemTableToSSTable(mem)
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	sst.fileptr.Close()

	// Reopen the file in read mode to test loading index
	sst, err = NewSSTable("test.sst", false, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to open SSTable: %v", err)
	}

	err = sst.LoadIndex()
	if err != nil {
		t.Fatalf("Failed to load index: %v", err)
	}

	if len(sst.Index) != 2 {
		t.Fatalf("Expected 2 keys in index, got %d", len(sst.Index))
	}
}
