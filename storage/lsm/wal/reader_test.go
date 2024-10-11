package wal

import (
	"os"
	"path/filepath"
	"testing"

	"universum/config"
	"universum/entity"
	"universum/storage/lsm/memtable"
)

func setupReaderTests(t *testing.T) {
	tmpdir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Storage.MaxRecordSizeInBytes = 1024
	config.Store.Storage.LSM.BloomFilterMaxRecords = 100
	config.Store.Storage.LSM.BloomFalsePositiveRate = 0.01
	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false
	config.Store.Storage.LSM.WriteBufferSize = 1024 * 1024
	config.Store.Logging.LogFileDirectory = tmpdir
}

func TestNewReader(t *testing.T) {
	setupReaderTests(t)
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
	setupReaderTests(t)
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	entries := []*WALRecord{
		{Key: "key1", Value: "value1", Expiry: 0, State: entity.RecordStateActive},
		{Key: "key1", Value: "value2", Expiry: 0, State: entity.RecordStateTombstoned},
	}

	ww, _ := NewWriter(dir)
	for _, entry := range entries {
		err := ww.AddToWALBuffer(entry.Key, entry.Value, entry.Expiry, entry.State)
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
		if entry.Key != expected.Key {
			t.Errorf("Entry %d mismatch: expected %+v, got %+v", i, expected, entry)
		}
		if entry.Value != expected.Value {
			t.Errorf("Entry %d value mismatch: expected %v, got %v", i, expected.Value, entry.Value)
		}
	}
}

func TestRestoreFromWAL(t *testing.T) {
	setupReaderTests(t)
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	entries := []*WALRecord{
		{Key: "key1", Value: "value1", Expiry: 0, State: entity.RecordStateActive},
		{Key: "key2", Value: "value2", Expiry: 0, State: entity.RecordStateTombstoned},
	}

	ww, _ := NewWriter(dir)
	for _, entry := range entries {
		err := ww.AddToWALBuffer(entry.Key, entry.Value, entry.Expiry, entry.State)
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

	record, code := memTable.Get("key1")
	if record == nil || code != entity.CRC_RECORD_FOUND {
		t.Errorf("Expected key1 to exist in memtable")
	}

	scalerRecord := record.(*entity.ScalarRecord)
	if scalerRecord.Value != "value1" {
		t.Errorf("Expected key1 to have value 'value1', got '%v'", scalerRecord.Value)
	}

	_, code = memTable.Get("key2")
	if code != entity.CRC_RECORD_TOMBSTONED {
		t.Errorf("Expected key2 to be tombstoned in memtable")
	}
}

func TestReaderClose(t *testing.T) {
	setupReaderTests(t)
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
	setupReaderTests(t)
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
