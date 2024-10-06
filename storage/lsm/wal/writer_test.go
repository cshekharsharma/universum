package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"universum/config"
)

func setupWriterTests() {
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false
	config.Store.Storage.LSM.WriteAheadLogBufferSize = 4096 // 4KB
	config.Store.Storage.LSM.WriteAheadLogFrequency = 1     // 1 second
	config.Store.Logging.LogFileDirectory = "/tmp"
}

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

func cleanupDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to clean up temp dir: %v", err)
	}
}

func TestNewWriter(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	if _, err := os.Stat(walFilePath); os.IsNotExist(err) {
		t.Fatalf("WAL file does not exist at path: %s", walFilePath)
	}
}

func TestAddToWALBufferSyncFlush(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	err = writer.AddToWALBuffer(OperationTypeSET, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after writing data")
	}
}

func TestAddToWALBufferAsyncFlush(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	err = writer.AddToWALBuffer("SET", "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after async flush")
	}
}

func TestFlush(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	err = writer.AddToWALBuffer("SET", "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	writer.flush()

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after flush")
	}
}

func TestBufferFlushOnMaxSize(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true
	config.Store.Storage.LSM.WriteAheadLogBufferSize = 100 // bytes

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	if writer.isFlushing {
		t.Fatalf("Expected isFlushing to be false initially")
	}

	numEntries := 3

	for i := 0; i < numEntries; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := writer.AddToWALBuffer("SET", key, value, 0)

		if err != nil {
			t.Fatalf("AddToWALBuffer failed: %v", err)
		}
	}

	time.Sleep(500 * time.Millisecond) // wait for flushing

	if writer.isFlushing {
		t.Errorf("Expected isFlushing to be false after flush")
	}

	if writer.buffer.Len() != 0 {
		t.Errorf("Expected buffer length to be 0 after flush, got %d", writer.buffer.Len())
	}

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after flush")
	}
}

func TestRotateWALFile(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	err = writer.AddToWALBuffer("SET", "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	err = writer.RotateWALFile()
	if err != nil {
		t.Fatalf("RotateWALFile failed: %v", err)
	}

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	fileInfo, err := os.Stat(walFilePath)
	if err != nil {
		t.Fatalf("Failed to stat WAL file: %v", err)
	}

	if fileInfo.Size() != 0 {
		t.Fatalf("WAL file size is not zero after rotation")
	}
}

func TestConcurrentAddToWALBuffer(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = false

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	numWritesPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numWritesPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				err := writer.AddToWALBuffer("SET", key, value, 0)
				if err != nil {
					t.Errorf("AddToWALBuffer failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after concurrent writes")
	}
}

func TestWriterClose(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}

	err = writer.AddToWALBuffer(OperationTypeDELETE, "key1", "", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	writer.Close()

	walFilePath := filepath.Join(dir, config.DefaultWALFileName)
	data, err := os.ReadFile(walFilePath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("WAL file is empty after Close")
	}
}

func TestGetEncodedEntries(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	err = writer.AddToWALBuffer("UNKNOWN_OP", "key1", "value1", 0)
	if err == nil {
		t.Fatalf("Expected error with unknown operation, but got nil")
	}

	err = writer.AddToWALBuffer(OperationTypeSET, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed with valid operation: %v", err)
	}
}

func TestAttemptFlushFailure(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}

	err = writer.AddToWALBuffer("SET", "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	writer.fileptr.Close()
	writer.flush()
}

func TestSyncCounterThreshold(t *testing.T) {
	setupWriterTests()
	dir := createTempDir(t)
	defer cleanupDir(t, dir)

	config.Store.Storage.LSM.WriteAheadLogAsyncFlush = true

	writer, err := NewWriter(dir)
	if err != nil {
		t.Fatalf("Failed to create WALWriter: %v", err)
	}
	defer writer.Close()

	writer.syncThreshold = 1

	err = writer.AddToWALBuffer("SET", "key1", "value1", 0)
	if err != nil {
		t.Fatalf("AddToWALBuffer failed: %v", err)
	}

	writer.flush()

	if writer.syncCounter != 0 {
		t.Errorf("Expected syncCounter to be reset to 0, got %d", writer.syncCounter)
	}
}
