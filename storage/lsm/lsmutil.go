package lsm

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"
	"universum/config"
)

const SstFileExtension = "sst"

func getAllSSTableFiles() ([]string, error) {
	dir := config.Store.Storage.LSM.DataStorageDirectory
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

func generateSSTableFileName() string {
	hasher := fnv.New64a()
	hasher.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	hash := hasher.Sum64()

	path := fmt.Sprintf("%s/%x.%s", config.Store.Storage.LSM.DataStorageDirectory, hash, SstFileExtension)
	return filepath.Clean(path)
}
