package filesys

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestAtomicCopyFileContent(t *testing.T) {
	srcFile, err := os.CreateTemp("", "src")
	if err != nil {
		t.Fatalf("Failed to create src file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	destFile, err := os.CreateTemp("", "dest")
	if err != nil {
		t.Fatalf("Failed to create dest file: %v", err)
	}
	defer os.Remove(destFile.Name())

	srcContent := []byte("This is a test.")
	if _, err := srcFile.Write(srcContent); err != nil {
		t.Fatalf("Failed to write to src file: %v", err)
	}
	srcFile.Close()

	if err := AtomicCopyFileContent(srcFile.Name(), destFile.Name()); err != nil {
		t.Errorf("Failed to copy file content atomically: %v", err)
	}

	destContent, err := os.ReadFile(destFile.Name())
	if err != nil {
		t.Errorf("Failed to read dest file: %v", err)
	}
	if string(destContent) != string(srcContent) {
		t.Errorf("Content mismatch: got %v, want %v", string(destContent), string(srcContent))
	}
}

func TestGetFileSizeInBytes_ValidFile(t *testing.T) {
	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name()) // Clean up

	data := []byte("Hello, World!")
	_, err = file.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	size, err := GetFileSizeInBytes(file)
	if err != nil {
		t.Errorf("GetFileSizeInBytes returned an error: %v", err)
	}

	expectedSize := int64(len(data))
	if size != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, size)
	}

	file.Close()
}

func TestGetFileSizeInBytes_ClosedFile(t *testing.T) {
	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	file.Close()
	defer os.Remove(file.Name()) // Clean up

	_, err = GetFileSizeInBytes(file)
	if err == nil {
		t.Errorf("Expected an error for closed file, but got none")
	}
}

func TestGetFileSizeInBytes_NilFilePointer(t *testing.T) {
	_, err := GetFileSizeInBytes(nil)
	if err == nil {
		t.Errorf("Expected an error for nil file pointer, but got none")
	}
}

func TestIsDirectoryWritable_WritableDirectory(t *testing.T) {
	dir := t.TempDir()

	if !IsDirectoryWritable(dir) {
		t.Errorf("Expected directory %s to be writable", dir)
	}
}

func TestIsDirectoryWritable_NonWritableDirectory(t *testing.T) {
	dir := t.TempDir()

	err := os.Chmod(dir, 0o500)
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}
	defer os.Chmod(dir, 0o700)

	if IsDirectoryWritable(dir) {
		t.Errorf("Expected directory %s to be non-writable", dir)
	}
}

func TestIsDirectoryWritable_NonExistentDirectory(t *testing.T) {
	dir := "/path/does/not/exist"

	if IsDirectoryWritable(dir) {
		t.Errorf("Expected non-existent directory %s to be non-writable", dir)
	}
}

func TestIsDiskFullError_ENOSPCError(t *testing.T) {
	err := &os.PathError{
		Op:   "write",
		Path: "/fake/path",
		Err:  syscall.ENOSPC,
	}

	if !IsDiskFullError(err) {
		t.Errorf("Expected true for ENOSPC error, got false")
	}
}

func TestIsDiskFullError_OtherError(t *testing.T) {
	err := &os.PathError{
		Op:   "write",
		Path: "/fake/path",
		Err:  syscall.EACCES,
	}

	if IsDiskFullError(err) {
		t.Errorf("Expected false for non-ENOSPC error, got true")
	}
}

func TestIsDiskFullError_NilError(t *testing.T) {
	if IsDiskFullError(nil) {
		t.Errorf("Expected false for nil error, got true")
	}
}

func TestIsDiskFullError_NonPathError(t *testing.T) {
	err := errors.New("some other error")

	if IsDiskFullError(err) {
		t.Errorf("Expected false for non-PathError, got true")
	}
}
