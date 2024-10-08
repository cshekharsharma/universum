package lsm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"universum/config"
)

func TestGetAllSSTableFiles(t *testing.T) {
	tempDir := t.TempDir()
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.DataStorageDirectory = tempDir

	files := []struct {
		name   string
		isSST  bool
		create bool
	}{
		{"file1.sst", true, true},
		{"file2.sst", true, true},
		{"file3.txt", false, true},
		{"file4", false, true},
		{"file5.sst", true, false},
	}

	for _, f := range files {
		if f.create {
			filePath := filepath.Join(tempDir, f.name)
			err := os.WriteFile(filePath, []byte("test data"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", f.name, err)
			}
		}
	}

	sstableFiles, err := getAllSSTableFiles()
	if err != nil {
		t.Fatalf("getAllSSTableFiles returned error: %v", err)
	}

	expectedFiles := map[string]bool{
		"file1.sst": true,
		"file2.sst": true,
	}

	if len(sstableFiles) != len(expectedFiles) {
		t.Fatalf("Expected %d SSTable files, got %d", len(expectedFiles), len(sstableFiles))
	}

	for _, f := range sstableFiles {
		if !expectedFiles[f] {
			t.Errorf("Unexpected SSTable file found: %s", f)
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
