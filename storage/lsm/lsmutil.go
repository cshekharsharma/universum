package lsm

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"universum/config"
	"universum/storage/lsm/sstable"
)

func getAllSSTableFiles() ([]string, error) {
	dir := config.Store.Storage.LSM.DataStorageDirectory

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSTable directory: %v", err)
	}

	var sstableFiles []string
	var sstExtWithDot = fmt.Sprintf(".%s", sstable.SstFileExtension)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), sstExtWithDot) {
			sstableFiles = append(sstableFiles, file.Name())
		}
	}

	sort.Slice(sstableFiles, func(i, j int) bool {
		timestampI := strings.TrimSuffix(sstableFiles[i], sstExtWithDot)
		timestampJ := strings.TrimSuffix(sstableFiles[j], sstExtWithDot)

		timeI, err1 := strconv.ParseInt(timestampI, 10, 64)
		timeJ, err2 := strconv.ParseInt(timestampJ, 10, 64)

		if err1 != nil || err2 != nil {
			return i < j
		}

		return timeI > timeJ
	})

	return sstableFiles, nil
}

func generateSSTableFileName() string {
	return fmt.Sprintf("%d.%s", time.Now().UnixNano(), sstable.SstFileExtension)
}
