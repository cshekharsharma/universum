package wal

import (
	"os"
	"path/filepath"
	"testing"

	"universum/config"
	"universum/entity"
	"universum/storage/lsm/memtable"
)

func setupReaderTests() {
	config.Store = config.GetSkeleton()
	config.Store.Storage.MaxRecordSizeInBytes = 1024
	config.Store.Storage.LSM.MaxMemtableRecords = 100
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false
	config.Store.Logging.LogFileDirectory = "/tmp"
}

func TestNewReader(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	file, err := os.Create(walFilePath)
	if err != nil {
		t.Fatalf("Failed to create WAL file: %v", err)
	}
	file.Close()

	reader, err := NewReader(dir)
	if err != nil {
		t.Fatalf("Failed to create WALReader: %v", err)
	}
	defer reader.Close()
}

func TestReadEntries(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	entries := []*WALRecord{
		{Operation: OperationTypeSET, Key: "key1", Value: "value1", TTL: 0},
		{Operation: OperationTypeDELETE, Key: "key2"},
	}

	ww, _ := NewWriter(dir)
	for _, entry := range entries {
		err := ww.AddToWALBuffer(entry.Operation, entry.Key, entry.Value, entry.TTL)
		if err != nil {
			t.Fatalf("Failed to write entry: %v", err)
		}
	}

	ww.Close()

	reader, err := NewReader(dir)
	if err != nil {
		t.Fatalf("Failed to create WALReader: %v", err)
	}
	defer reader.Close()

	readEntries, err := reader.readEntries()
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	if len(readEntries) != len(entries) {
		t.Fatalf("Expected %d entries, got %d", len(entries), len(readEntries))
	}

	for i, entry := range readEntries {
		expected := entries[i]
		if entry.Operation != expected.Operation || entry.Key != expected.Key {
			t.Errorf("Entry %d mismatch: expected %+v, got %+v", i, expected, entry)
		}
		if entry.Operation == OperationTypeSET && entry.Value != expected.Value {
			t.Errorf("Entry %d value mismatch: expected %v, got %v", i, expected.Value, entry.Value)
		}
	}
}

func TestRestoreFromWAL(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	entries := []*WALRecord{
		{Operation: OperationTypeSET, Key: "key1", Value: "value1", TTL: 0},
		{Operation: OperationTypeSET, Key: "key2", Value: "value2", TTL: 0},
		{Operation: OperationTypeDELETE, Key: "key1"},
	}

	ww, _ := NewWriter(dir)
	for _, entry := range entries {
		err := ww.AddToWALBuffer(entry.Operation, entry.Key, entry.Value, entry.TTL)
		if err != nil {
			t.Fatalf("Failed to write entry: %v", err)
		}
	}

	ww.Close()

	reader, err := NewReader(dir)
	if err != nil {
		t.Fatalf("Failed to create WALReader: %v", err)
	}
	defer reader.Close()

	memTable := memtable.CreateNewMemTable(config.MemtableStorageTypeLB)

	keycount, err := reader.RestoreFromWAL(memTable)
	if keycount != int64(len(entries)) {
		t.Errorf("Expected %d keys to be restored, got %d", len(entries), keycount)
	}

	if err != nil {
		t.Fatalf("Failed to restore from WAL: %v", err)
	}

	exists, _ := memTable.Exists("key1")
	if exists {
		t.Errorf("Expected key1 to be deleted from memtable")
	}

	record, code := memTable.Get("key2")
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Errorf("Expected key2 to exist in memtable")
	}

	scalerRecord := record.(*entity.ScalarRecord)
	if scalerRecord.Value != "value2" {
		t.Errorf("Expected key2 to have value 'value2', got '%v'", scalerRecord.Value)
	}
}

func TestReaderClose(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	file, err := os.Create(walFilePath)
	if err != nil {
		t.Fatalf("Failed to create WAL file: %v", err)
	}
	file.Close()

	reader, err := NewReader(dir)
	if err != nil {
		t.Fatalf("Failed to create WALReader: %v", err)
	}

	reader.Close()

	_, err = reader.readEntries()
	if err == nil {
		t.Errorf("Expected error when reading after close, but got nil")
	}
}

func TestReadEntriesWithCorruptedData(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	err := os.WriteFile(walFilePath, []byte{0x00, 0x01, 0x02}, 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted data to WAL file: %v", err)
	}

	reader, err := NewReader(dir)
	if err != nil {
		t.Fatalf("Failed to create WALReader: %v", err)
	}
	defer reader.Close()

	_, err = reader.readEntries()
	if err == nil {
		t.Errorf("Expected error when reading corrupted data, but got nil")
	}
}

func TestRestoreFromWALWithUnknownOperation(t *testing.T) {
	setupReaderTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	walRecord := &WALRecord{Operation: "UnknownOp", Key: "key1", Value: "value1", TTL: 0}

	ww, _ := NewWriter(dir)
	err := ww.AddToWALBuffer(walRecord.Operation, walRecord.Key, walRecord.Value, walRecord.TTL)
	if err == nil {
		t.Fatalf("Expected error with unknown operation, but got nil")
	}
}
