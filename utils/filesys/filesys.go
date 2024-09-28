package filesys

import (
	"errors"
	"io"
	"os"
	"syscall"
)

func AtomicCopyFileContent(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	tmpFileName := dest + ".tmp"
	tmpFile, err := os.Create(tmpFileName)

	if err != nil {
		return err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		os.Remove(tmpFileName)
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		os.Remove(tmpFileName)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFileName)
		return err
	}

	if err := os.Rename(tmpFileName, dest); err != nil {
		os.Remove(tmpFileName)
		return err
	}

	return nil
}

// GetFileSizeInBytes returns the size of the file in bytes.
// It takes a file pointer as input and returns the file size in bytes.
// If an error occurs, it returns the error.
func GetFileSizeInBytes(filePtr *os.File) (int64, error) {
	fileInfo, err := filePtr.Stat()
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

// IsDirectoryWritable checks if a directory is writable.
// It takes the directory path as input and returns a boolean value.
// If the directory is writable, it returns true; otherwise, it returns false.
func IsDirectoryWritable(dirPath string) bool {
	tempFile, err := os.CreateTemp(dirPath, "writetest")
	if err != nil {
		return false
	}

	tempFile.Close()
	os.Remove(tempFile.Name())

	return true
}

// isDiskFullError checks if the error is caused by a lack of disk space
func IsDiskFullError(err error) bool {
	if errno, ok := err.(*os.PathError); ok {
		return errors.Is(errno.Err, syscall.ENOSPC)
	}
	return false
}
