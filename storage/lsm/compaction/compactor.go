package compaction

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/storage/lsm/sstable"
)

const (
	DefaultMaxLevel             int = 8 // Default maximum number of levels in LSM
	CompactionThresholdPerLevel int = 3 // Number of SSTables per level before triggering compaction
)

type Compactor struct {
	LevelSSTables map[int64][]*sstable.SSTable // SSTables grouped by level
	compactionMu  sync.Mutex                   // Concurrency safety for compacting
	additionMu    sync.Mutex                   // Concurrency safety for adding SSTables
	MaxLevel      int64                        // Maximum number of levels
}

func NewCompactor() *Compactor {
	return &Compactor{
		LevelSSTables: make(map[int64][]*sstable.SSTable),
		MaxLevel:      int64(DefaultMaxLevel),
	}
}

func (c *Compactor) AddSSTable(level int64, sst *sstable.SSTable) {
	c.additionMu.Lock()
	defer c.additionMu.Unlock()

	c.LevelSSTables[level] = append(c.LevelSSTables[level], sst)
}

func (c *Compactor) Compact() {
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Error("SSTable Compactor: Recovered from panic: %v", r)
			go c.Compact() // Restart the compaction process if it panics.
		}
	}()

	for {
		for level := int64(0); level < c.MaxLevel; level++ {
			if len(c.LevelSSTables[level]) >= int(CompactionThresholdPerLevel) {
				c.CompactLevel(level)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *Compactor) CompactLevel(level int64) error {
	c.compactionMu.Lock()
	defer c.compactionMu.Unlock()

	if _, ok := c.LevelSSTables[level]; !ok {
		return nil // No SSTables to compact
	}

	sstablesToCompact := c.LevelSSTables[level]
	if len(sstablesToCompact) < int(CompactionThresholdPerLevel) {
		return nil // Not enough SSTables to compact
	}

	overlappingSSTs := c.getOverlappingSSTables(level+1, sstablesToCompact)
	mergedSST, err := c.mergeSSTables(sstablesToCompact, overlappingSSTs)
	if err != nil {
		logger.Get().Error("SSTable compaction failed: %v", err)
		return err
	}

	SSTReplacementChan <- &SSTReplacement{
		Obsoletes:  sstablesToCompact,
		Substitute: mergedSST,
	}

	time.Sleep(10 * time.Microsecond) // give consumer some time to process the notification

	c.LevelSSTables[level] = nil

	if level+1 >= c.MaxLevel {
		c.AddSSTable(level, mergedSST) // cannot go above maxlevel, so add at the same level
	} else {
		c.AddSSTable(level+1, mergedSST)
	}

	err = c.deleteOldSSTables(sstablesToCompact)
	if err != nil {
		logger.Get().Error("Failed to clean obsolete sstables post compaction: %v", err)
		return err
	}

	return nil
}

func (c *Compactor) mergeSSTables(sstables, overlapping []*sstable.SSTable) (*sstable.SSTable, error) {
	allSSTables := append(sstables, overlapping...)
	mergedList := make([]*entity.RecordKV, 0)

	for _, sstable := range allSSTables {
		sstRecords, err := sstable.GetAllRecords()
		if err != nil {
			return nil, fmt.Errorf("failed to get all records from SSTable: %v", err)
		}

		mergedList = Merge(mergedList, sstRecords)
	}

	compactedMergedList := make([]*entity.RecordKV, 0)
	for _, recordKV := range mergedList {
		if recordKV.Record.IsExpired() || recordKV.Record.IsTombstoned() {
			continue // skip all expired and deleted/tombstoned records
		}
		compactedMergedList = append(compactedMergedList, recordKV)
	}

	newSSTFileName, err := c.getMergedSSTFileName(sstables)
	if err != nil {
		return nil, err
	}

	mergedSST, err := sstable.NewSSTable(
		newSSTFileName,
		sstable.SSTmodeWrite,
		config.Store.Storage.LSM.BloomFilterMaxRecords,
		config.Store.Storage.LSM.BloomFalsePositiveRate,
	)

	if err != nil {
		return nil, err
	}

	err = mergedSST.FlushRecordsToSSTable(compactedMergedList)
	if err != nil {
		return nil, err
	}

	return mergedSST, nil
}

func (c *Compactor) getMergedSSTFileName(sstables []*sstable.SSTable) (string, error) {
	var min int64 = 1<<63 - 1

	for _, sst := range sstables {
		filenamestr := strings.TrimSuffix(filepath.Base(sst.Filename), filepath.Ext(filepath.Base(sst.Filename)))
		filenameInt, err := strconv.ParseInt(filenamestr, 10, 64)

		if err != nil {
			return "", fmt.Errorf("failed to parse SST filename %s to int64: %v", sst.Filename, err)
		}

		if filenameInt < min {
			min = filenameInt
		}
	}

	return fmt.Sprintf("%d.%s", min, sstable.SstFileExtension), nil
}

func (c *Compactor) getOverlappingSSTables(nextLevel int64, sstables []*sstable.SSTable) []*sstable.SSTable {
	overlappingSSTables := []*sstable.SSTable{}

	if _, ok := c.LevelSSTables[nextLevel]; !ok {
		return overlappingSSTables // No SSTables at the next level
	}

	nextLevelSSTables := c.LevelSSTables[nextLevel]

	for _, sst := range sstables {
		for _, nextSST := range nextLevelSSTables {
			if c.doesOverlap(sst, nextSST) {
				overlappingSSTables = append(overlappingSSTables, nextSST)
			}
		}
	}

	return overlappingSSTables
}

func (c *Compactor) doesOverlap(sst1, sst2 *sstable.SSTable) bool {
	sst1FirstKey := sst1.Metadata.FirstKey
	sst1LastKey := sst1.Metadata.LastKey
	sst2FirstKey := sst2.Metadata.FirstKey
	sst2LastKey := sst2.Metadata.LastKey

	return !(sst1LastKey < sst2FirstKey || sst2LastKey < sst1FirstKey)
}

func (c *Compactor) deleteOldSSTables(sstables []*sstable.SSTable) error {
	for _, sst := range sstables {
		err := sst.DeleteFromDisk()
		if err != nil {
			return err
		}
	}
	return nil
}
