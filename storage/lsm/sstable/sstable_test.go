package sstable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"universum/config"
	"universum/entity"
	"universum/storage/lsm/memtable"
	"universum/utils"
)

func SetUpSSTableTests(t *testing.T) {
	tmpdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Logging.LogFileDirectory = tmpdir
	config.Store.Storage.StorageEngine = config.StorageEngineLSM
	config.Store.Storage.MaxRecordSizeInBytes = 1048576
	config.Store.Storage.LSM.MemtableStorageType = config.MemtableStorageTypeLB
	config.Store.Storage.LSM.BloomFilterMaxRecords = 1000
	config.Store.Storage.LSM.WriteBufferSize = 1048576
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteBlockSize = 100
	config.Store.Storage.LSM.DataStorageDirectory = tmpdir
	config.Store.Storage.LSM.BlockCompressionAlgo = config.CompressionAlgoLZ4
}

func TestNewSSTable(t *testing.T) {
	SetUpSSTableTests(t)

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, SSTmodeWrite, lsmCnf.BloomFilterMaxRecords, lsmCnf.BloomFalsePositiveRate)
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
	SetUpSSTableTests(t)

	fileName := "test1.sst"
	cnf := config.Store.Storage.LSM

	sst, err := NewSSTable(fileName, SSTmodeWrite, 100, 0.01)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", cnf.DataStorageDirectory, fileName)))

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("value%d", i)
		mem.Set(key, val, int64(i*1000), entity.RecordStateActive)
	}

	err = sst.FlushRecordsToSSTable(mem.GetAll())
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	// reset sst instance to populate again
	sst, _ = NewSSTable(fileName, SSTmodeRead, 100, 0.01)

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
	SetUpSSTableTests(t)

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, SSTmodeWrite, lsmCnf.BloomFilterMaxRecords, lsmCnf.BloomFalsePositiveRate)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath)

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 100, entity.RecordStateActive)
	mem.Set("key2", "value2", 100, entity.RecordStateActive)

	err = sst.FlushRecordsToSSTable(mem.GetAll())
	if err != nil {
		t.Fatalf("Failed to flush memtable to SSTable: %v", err)
	}

	if sst.RecordCount != 2 {
		t.Fatalf("Expected 2 records in SSTable, got %d", sst.RecordCount)
	}

	if sst.Metadata.DataSize < 676 || sst.Metadata.DataSize > 684 {
		t.Fatalf("Expected 676B of size in SSTable metadata, got %dB", sst.Metadata.DataSize)
	}

	if sst.DataSize < 775 && sst.DataSize > 784 {
		t.Fatalf("Expected 775B of data in SSTable, got %dB", sst.DataSize)
	}

	_, err = os.Stat(filepath)
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("SSTable file does not exist at %s", filepath)
	}
}

func TestLoadBlockAndFindRecord(t *testing.T) {
	SetUpSSTableTests(t)

	filename := "test.sst"
	filepath := filepath.Clean(fmt.Sprintf(
		"%s/%s", config.Store.Storage.LSM.DataStorageDirectory, filename))

	lsmCnf := config.Store.Storage.LSM
	sst, err := NewSSTable(filename, SSTmodeWrite, lsmCnf.BloomFilterMaxRecords, lsmCnf.BloomFalsePositiveRate)
	if err != nil {
		t.Fatalf("Failed to create SSTable: %v", err)
	}
	defer os.Remove(filepath)

	mem := memtable.CreateNewMemTable(config.DefaultMemtableStorageType).(*memtable.ListBloomMemTable)
	mem.Set("key1", "value1", 10, entity.RecordStateActive)
	mem.Set("key2", "value2", 10, entity.RecordStateActive)

	_ = sst.FlushRecordsToSSTable(mem.GetAll())

	entry, _ := sst.FindBlockForKey("key1", sst.Index)
	blockOffset, blocksize := utils.UnpackNumbers(entry.GetOffset())

	block, err := sst.LoadBlock(int64(blockOffset), int64(blocksize))
	if err != nil {
		t.Fatalf("Failed to read block: %v", err)
	}

	recordList, err := block.GetAllRecords()
	if err != nil {
		t.Fatalf("Failed to read all sstable records: %v", err)
	}

	if len(recordList) == 0 {
		t.Fatalf("Expected records in the block, found 0")
	}

	if len(block.Data) == 0 {
		t.Fatalf("Expected data in the block, found empty")
	}

	offset, _ := block.Index.Load("key1")
	key, value, err := block.ReadRecordAtOffset(offset.(int64))

	if err != nil {
		t.Fatalf("Failed to read record from block: %v", err)
	}

	if string(key) != "key1" {
		t.Fatalf("Expected key1, got %s", key)
	}

	if value == nil {
		t.Fatalf("Expected value1, got %s", value)
	}

	found, _, err := sst.FindRecord("key1")
	if !found {
		t.Fatalf("Expected record to be found")
	}

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	found, record, err := sst.FindRecord("key1")
	if !found {
		t.Fatalf("Expected record to be found")
	}

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var expectedVal interface{} = "value1"
	if record.GetValue() != expectedVal {
		t.Fatalf("Expected value1, got %s", record.GetValue())
	}

	found, _, _ = sst.FindRecord("key3")
	if found {
		t.Fatalf("Expected record to not be found")
	}

	_, _, err = sst.FindRecord("")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	allrecords, err := sst.GetAllRecords()
	if err != nil {
		t.Fatalf("GetAllRecords: expected nil but got error: %v", err)
	}

	if len(allrecords) != 2 {
		t.Fatalf("GetAllRecords: expected 2 records, got %d", len(allrecords))
	}

	if allrecords[0].Key != "key1" || allrecords[0].Record.GetValue() != "value1" {
		t.Fatalf("GetAllRecords: expected key1=value1, got %v", allrecords[0])
	}

	if allrecords[1].Key != "key2" || allrecords[1].Record.GetValue() != "value2" {
		t.Fatalf("GetAllRecords: expected key2=value1, got %v", allrecords[1])
	}
}
