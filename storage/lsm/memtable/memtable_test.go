package memtable

import (
	"testing"
	"universum/config"
)

func TestCreateNewMemTable_TypeLB(t *testing.T) {
	tmpdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.BloomFilterMaxRecords = 1000
	config.Store.Storage.LSM.WriteBufferSize = 1048576
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Logging.LogFileDirectory = tmpdir

	memTable := CreateNewMemTable(config.MemtableStorageTypeLB)
	_, ok := memTable.(*ListBloomMemTable)
	if !ok {
		t.Errorf("Expected memTable to be of type *ListBloomMemTable, got %T", memTable)
	}
}

func TestCreateNewMemTable_TypeTB(t *testing.T) {
	config.Store = config.GetSkeleton()

	memTable := CreateNewMemTable(config.MemtableStorageTypeTB)
	_, ok := memTable.(*TreeBloomMemTable)
	if !ok {
		t.Errorf("Expected memTable to be of type *ListMapMemTable, got %T", memTable)
	}
}

func TestCreateNewMemTable_Default(t *testing.T) {
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.BloomFilterMaxRecords = 1000
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01

	memTable := CreateNewMemTable("randomvalue")
	_, ok := memTable.(*ListBloomMemTable)
	if !ok {
		t.Errorf("Expected memTable to be of type *ListBloomMemTable, got %T", memTable)
	}
}
