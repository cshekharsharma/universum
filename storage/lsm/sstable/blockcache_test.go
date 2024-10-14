package sstable

import (
	"testing"
	"time"
	"universum/config"
	"universum/dslib"
	"universum/entity"
)

func setupBlockCacheTests() {
	config.Store = config.GetSkeleton()
	config.Store.Storage.LSM.WriteBlockSize = 4096
	config.Store.Storage.LSM.BlockCacheMemoryLimit = 22500
}

func createTestBlock(keys []string) *Block {
	expiry := time.Now().Unix() + 10000
	bloom := dslib.NewBloomFilter(1000, 5)

	block := NewBlock(1024)
	for i := 0; i < len(keys); i++ {
		block.AddRecord(keys[i], map[string]interface{}{
			"Value":  int64(100),
			"LAT":    0,
			"Expiry": expiry,
			"State":  0,
		},
			bloom)
	}

	block.SerializeBlock()
	block.SetID(GenerateBlockID(block.FirstKey, block.LastKey))
	return block
}

func TestBlockCacheAdd(t *testing.T) {
	setupBlockCacheTests()
	blockCache := NewBlockCache()
	block := createTestBlock([]string{"key1", "key2", "key3"})
	block2 := createTestBlock([]string{"key1", "key2", "key3"})

	blockCache.Add(block)
	blockCache.Add(block2)

	cachedBlock, found := blockCache.GetBlock(block.GetID())
	if !found {
		t.Fatalf("Block with ID %d was not found in cache", block.GetID())
	}
	if cachedBlock != block {
		t.Fatalf("Expected block and cached block to be the same")
	}
}

func TestBlockCacheEviction(t *testing.T) {
	setupBlockCacheTests()
	blockCache := NewBlockCache()
	block1 := createTestBlock([]string{"key1", "key2"})
	block2 := createTestBlock([]string{"firstkey", "keyA"})
	block3 := createTestBlock([]string{"samplekey", "validsearch"})

	blockCache.Add(block1)
	blockCache.Add(block2)
	blockCache.Add(block3)

	_, found := blockCache.GetBlock(block1.Id)
	if found {
		t.Fatalf("Block with ID %d should have been evicted", block1.Id)
	}

	_, found = blockCache.GetBlock(block2.Id)
	if !found {
		t.Fatalf("Block with ID %d should still be in cache", block2.Id)
	}
	_, found = blockCache.GetBlock(block3.Id)
	if !found {
		t.Fatalf("Block with ID %d should still be in cache", block3.Id)
	}
}

func TestBlockCacheGet(t *testing.T) {
	setupBlockCacheTests()
	blockCache := NewBlockCache()
	block := createTestBlock([]string{"key1", "key2"})

	blockCache.Add(block)

	cachedBlock, found := blockCache.GetBlock(block.Id)
	if !found {
		t.Fatalf("Block with ID %d was not found in cache", block.Id)
	}
	if cachedBlock != block {
		t.Fatalf("Expected block and cached block to be the same")
	}
}

func TestBlockCacheSearchBlock(t *testing.T) {
	setupBlockCacheTests()
	blockCache := NewBlockCache()
	block := createTestBlock([]string{"key1", "key2"})

	blockCache.Add(block)

	record, err := blockCache.SearchBlock(block.Id, "key1")
	if err != nil {
		t.Fatalf("Error occurred during block search: %v", err)
	}

	scalarRecord := record.(*entity.ScalarRecord)
	if scalarRecord.Value != int64(100) {
		t.Fatalf("Expected 100, got %v", scalarRecord.Value)
	}
}

func TestBlockCacheShardDistribution(t *testing.T) {
	setupBlockCacheTests()
	blockCache := NewBlockCache()
	block1 := createTestBlock([]string{"key1", "key2"})
	block2 := createTestBlock([]string{"key11", "key22", "key33"})

	blockCache.Add(block1)
	blockCache.Add(block2)

	shard1 := blockCache.shardForBlockID(block1.Id)
	if _, found := shard1.cache.Load(block1.Id); !found {
		t.Fatalf("Block with ID %d should be in shard1", block1.Id)
	}

	shard2 := blockCache.shardForBlockID(block2.Id)
	if _, found := shard2.cache.Load(block2.Id); !found {
		t.Fatalf("Block with ID %d should be in shard2", block2.Id)
	}
}
