package cache

import (
	"container/list"
	"errors"
	"sync"
	"universum/config"
	"universum/entity"
	"universum/storage/lsm/sstable"
)

const (
	ShardCount int = 64 // Number of shards
)

type BlockCacheShard struct {
	cache       sync.Map   // map[uint64]*list.Element
	eviction    *list.List // Doubly linked list for LRU eviction
	currentSize int64      // Current size of the shard's cache
}

type BlockCache struct {
	shards [ShardCount]*BlockCacheShard
}

// CacheItem represents an item in the cache, stored in the eviction list
type CacheItem struct {
	BlockID   uint64
	BlockData *sstable.Block
}

func NewBlockCache() *BlockCache {
	bc := &BlockCache{}
	for i := 0; i < int(ShardCount); i++ {
		bc.shards[i] = &BlockCacheShard{
			cache:    sync.Map{},
			eviction: list.New(),
		}
	}
	return bc
}

func (bc *BlockCache) shardForBlockID(blockID uint64) *BlockCacheShard {
	return bc.shards[blockID%uint64(ShardCount)]
}

func (bc *BlockCache) GetBlock(blockID uint64) (*sstable.Block, bool) {
	shard := bc.shardForBlockID(blockID)

	if element, found := shard.cache.Load(blockID); found {
		shard.eviction.MoveToFront(element.(*list.Element))
		return element.(*list.Element).Value.(*CacheItem).BlockData, true
	}
	return nil, false
}

func (bc *BlockCache) Add(block *sstable.Block) {
	blockID := block.GetID()
	shard := bc.shardForBlockID(blockID)

	if element, found := shard.cache.Load(blockID); found {
		shard.eviction.MoveToFront(element.(*list.Element))
		return
	}

	maxCacheSize := config.Store.Storage.LSM.BlockCacheMemoryLimit
	counter := 10
	for shard.currentSize+block.CurrentSize > maxCacheSize/int64(ShardCount) && counter > 0 {
		shard.evict()
		counter--
	}

	item := &CacheItem{
		BlockID:   blockID,
		BlockData: block,
	}

	element := shard.eviction.PushFront(item)
	shard.cache.Store(blockID, element)
	shard.currentSize += block.CurrentSize
}

func (shard *BlockCacheShard) evict() {
	element := shard.eviction.Back()
	if element != nil {
		cacheItem := element.Value.(*CacheItem)
		shard.cache.Delete(cacheItem.BlockID)
		shard.eviction.Remove(element)
		shard.currentSize -= cacheItem.BlockData.CurrentSize
	}
}

func (bc *BlockCache) SearchBlock(sst *sstable.SSTable, blockID uint64, key string) (entity.Record, error) {
	block, found := bc.GetBlock(blockID)
	if found {
		return bc.searchInBlock(block, key)
	}
	return nil, errors.New("cache miss: block could not be found in the block cache")
}

// searchInBlock searches for a key in a given block's data
func (bc *BlockCache) searchInBlock(block *sstable.Block, key string) (entity.Record, error) {
	record, err := block.GetRecord(key)
	if err != nil {
		return nil, err
	}
	return record, nil
}
