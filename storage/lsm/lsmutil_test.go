package lsm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"universum/config"
)

func TestGetAllSSTableFilesWithTempDir(t *testing.T) {
	tempDir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.DataStorageDirectory = tempDir

	timeStamps := []string{
		fmt.Sprintf("1625151601234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151603234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151604234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151605234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151602234567890.%s", SstFileExtension),
	}

	for _, ts := range timeStamps {
		filePath := filepath.Join(tempDir, ts)
		_, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", ts, err)
		}
	}

	retreivedFiles, err := getAllSSTableFiles()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedOrder := []string{
		fmt.Sprintf("1625151605234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151604234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151603234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151602234567890.%s", SstFileExtension),
		fmt.Sprintf("1625151601234567890.%s", SstFileExtension),
	}

	if len(retreivedFiles) != len(expectedOrder) {
		t.Fatalf("Expected %d files, got %d", len(expectedOrder), len(retreivedFiles))
	}

	for i, expectedFileName := range expectedOrder {
		if retreivedFiles[i] != expectedFileName {
			t.Errorf("Expected file %s at index %d, got %s", expectedFileName, i, retreivedFiles[i])
		}
	}
}

func TestGetAllSSTableFilesEmptyDirectory(t *testing.T) {
	config.Store = config.GetSkeleton()
	tempDir := t.TempDir()
	config.Store.Storage.LSM.DataStorageDirectory = tempDir

	sstableFiles, err := getAllSSTableFiles()
	if err != nil {
		t.Fatalf("getAllSSTableFiles returned error: %v", err)
	}

	if len(sstableFiles) != 0 {
		t.Errorf("Expected 0 SSTable files, got %d", len(sstableFiles))
	}
}

func TestGetAllSSTableFilesDirectoryNotExist(t *testing.T) {
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.DataStorageDirectory = "/non/existent/directory"

	_, err := getAllSSTableFiles()
	if err == nil {
		t.Fatalf("Expected error when directory does not exist, got nil")
	}
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "failed to read SSTable directory") {
		t.Errorf("Expected directory not exist error, got: %v", err)
	}
}

func TestGenerateSSTableFileName(t *testing.T) {
	config.Store = config.GetSkeleton()
	tempDir := t.TempDir()
	config.Store.Storage.LSM.DataStorageDirectory = tempDir

	fileName1 := generateSSTableFileName()

	if !strings.HasSuffix(fileName1, SstFileExtension) {
		t.Errorf("Expected file name to have extension %s, got %s", SstFileExtension, fileName1)
	}

	nameParts := strings.Split(fileName1, ".")
	if len(nameParts) != 2 {
		t.Errorf("Expected file name to have format 'hash.%s', got %s", SstFileExtension, fileName1)
	}

	time.Sleep(1 * time.Nanosecond)
	fileName2 := generateSSTableFileName()
	if fileName1 == fileName2 {
		t.Errorf("Expected different file names on subsequent calls, got %s and %s", fileName1, fileName2)
	}
}
