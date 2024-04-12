package filesys

import (
	"io"
	"os"
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
