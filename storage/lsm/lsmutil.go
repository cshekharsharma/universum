package lsm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"universum/config"
)

const SstFileExtension = ".sst"

func getAllSSTableFiles() ([]string, error) {
	dir := config.GetDataStoragePath()
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSTable directory: %v", err)
	}

	var sstableFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), SstFileExtension) {
			sstableFiles = append(sstableFiles, filepath.Join(dir, file.Name()))
		}
	}
	return sstableFiles, nil
}
