package lsm

import (
	"fmt"
	"hash/fnv"
	"os"
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
			sstableFiles = append(sstableFiles, file.Name())
		}
	}
	return sstableFiles, nil
}

func generateSSTableFileName() string {
	hasher := fnv.New64a()
	hasher.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	hash := hasher.Sum64()

	return fmt.Sprintf("%x.%s", hash, SstFileExtension)
}
