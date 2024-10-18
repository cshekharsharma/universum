package compaction

import (
	"fmt"
	"sync"
	"time"
	"universum/internal/logger"
	"universum/storage/lsm/sstable"
)

const (
	DefaultMaxLevel             int = 6 // Default maximum number of levels in LSM
	CompactionThresholdPerLevel int = 4 // Number of SSTables per level before triggering compaction
)

type Compactor struct {
	LevelSSTables map[int64][]*sstable.SSTable // SSTables grouped by level
	mutex         sync.Mutex                   // Concurrency safety for compacting
	MaxLevel      int64                        // Maximum number of levels
}

func NewCompactor() *Compactor {
	return &Compactor{
		LevelSSTables: make(map[int64][]*sstable.SSTable),
		MaxLevel:      int64(DefaultMaxLevel),
	}
}

func (c *Compactor) AddSSTable(level int64, sst *sstable.SSTable) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.LevelSSTables[level] = append(c.LevelSSTables[level], sst)
}

func (c *Compactor) Compact() {
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Error("SSTable Compactor: Recovered from panic: %v", r)
			go c.Compact() // Restart the compaction process if it panics.
		}
	}()

	for level := int64(0); level < c.MaxLevel; level++ {
		if len(c.LevelSSTables[level]) > int(CompactionThresholdPerLevel) {
			c.CompactLevel(level)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Compactor) CompactLevel(level int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.LevelSSTables[level]; !ok {
		return nil // No SSTables to compact
	}

	sstablesToCompact := c.LevelSSTables[level]
	if len(sstablesToCompact) < int(CompactionThresholdPerLevel) {
		return nil // Not enough SSTables to compact
	}

	mergedSST, err := c.mergeSSTables(sstablesToCompact)
	if err != nil {
		logger.Get().Error("SSTable compactopn failed: %v", err)
		return err
	}

	c.LevelSSTables[level] = nil
	c.AddSSTable(level+1, mergedSST)

	err = c.deleteOldSSTables(sstablesToCompact)
	if err != nil {
		logger.Get().Error("Failed to clean obsolete sstables post compaction: %v", err)
		return err
	}

	return nil
}

func (c *Compactor) mergeSSTables(sstables []*sstable.SSTable) (*sstable.SSTable, error) {
	for i, sstable := range sstables {
		fmt.Printf("merge sstable at key:%d, %v", i, sstable)
	}
	return nil, nil
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
