package sstable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	"universum/config"
	"universum/entity"
	"universum/storage/lsm/memtable"
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

	err = sst.FlushMemTableToSSTable(mem)
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	if sst.RecordCount != 2 {
		t.Fatalf("Expected 2 records in SSTable, got %d", sst.RecordCount)
	}

	blockOffset := sst.Index["key1"]
	block, err := sst.ReadBlock(blockOffset)
	if err != nil {
		t.Fatalf("Failed to read block: %v", err)
	}

	if len(block.records) == 0 {
		t.Fatalf("Expected records in the block")
	}
}

func TestFlushIndex(t *testing.T) {
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

	err = sst.FlushIndex()
	if err != nil {
		t.Fatalf("Failed to flush index: %v", err)
	}

	if len(sst.Index) != 2 {
		t.Fatalf("Expected 2 keys in index, got %d", len(sst.Index))
	}
}

func TestWriteMetadata(t *testing.T) {
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

	if sst.Metadata.NumRecords != 1 {
		t.Fatalf("Expected 1 record in metadata, got %d", sst.Metadata.NumRecords)
	}
}

func TestLoadBloomFilter(t *testing.T) {
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

	// Reopen SSTable in read mode to simulate loading bloom filter
	sst, err = NewSSTable("test.sst", false, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to open SSTable: %v", err)
	}

	err = sst.LoadBloomFilter()
	if err != nil {
		t.Fatalf("Failed to load bloom filter: %v", err)
	}

	if !sst.BloomFilter.Exists("key1") {
		t.Fatalf("Expected bloom filter to contain key1")
	}
}

func TestFlushBlock(t *testing.T) {
	sst, err := NewSSTable("test.sst", true, 1000, 0.01)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove("test.sst")

	mem := memtable.CreateNewMemTable("skiplist").(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 0)

	record := &entity.ScalarRecord{
		Value:  "value1",
		Expiry: time.Now().Unix() + 1800,
	}

	err = sst.CurrentBlock.AddRecord("key1", record.ToMap(), sst.BloomFilter)
	if err != nil {
		t.Fatalf("Failed to add record to block: %v", err)
	}

	err = sst.FlushBlock()
	if err != nil {
		t.Fatalf("Failed to flush block: %v", err)
	}

	if len(sst.Index) == 0 {
		t.Fatalf("Expected block to be flushed and indexed, but index is empty")
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

func TestFlushBloomFilter(t *testing.T) {
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

	err = sst.FlushBloomFilter()
	if err != nil {
		t.Fatalf("Failed to flush bloom filter: %v", err)
	}

	if sst.Metadata.BloomFilterSize == 0 {
		t.Fatalf("Expected BloomFilter size to be set, but it is 0")
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
