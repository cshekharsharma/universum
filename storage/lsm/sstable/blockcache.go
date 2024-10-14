package sstable

import (
	"container/list"
	"errors"
	"sync"
	"universum/config"
	"universum/entity"
)

const (
	ShardCount               int = 1 << 6 // 64 shards
	maxEvictionRetriesIfFull int = 1 << 3 // 8 retries
)

var BlockCacheStore *BlockCache

type BlockCacheShard struct {
	cache       sync.Map   // map[uint64]*list.Element
	eviction    *list.List // Doubly linked list for LRU eviction
	maxsize     int64      // Max size of the shard's cache
	currentSize int64      // Current size of the shard's cache
}

type BlockCache struct {
	shards [ShardCount]*BlockCacheShard
}

type CacheItem struct {
	BlockID   uint64
	BlockData *Block
}

func NewBlockCache() *BlockCache {
	bc := &BlockCache{}
	for i := 0; i < int(ShardCount); i++ {
		bc.shards[i] = &BlockCacheShard{
			cache:       sync.Map{},
			eviction:    list.New(),
			maxsize:     config.Store.Storage.LSM.BlockCacheMemoryLimit / int64(ShardCount),
			currentSize: 0,
		}
	}
	return bc
}

func (bc *BlockCache) shardForBlockID(blockID uint64) *BlockCacheShard {
	return bc.shards[blockID%uint64(ShardCount)]
}

func (bc *BlockCache) GetBlock(blockID uint64) (*Block, bool) {
	shard := bc.shardForBlockID(blockID)

	if element, found := shard.cache.Load(blockID); found {
		shard.eviction.MoveToFront(element.(*list.Element))
		return element.(*list.Element).Value.(*CacheItem).BlockData, true
	}
	return nil, false
}

func (bc *BlockCache) Add(block *Block) {
	blockID := block.GetID()
	shard := bc.shardForBlockID(blockID)

	if element, found := shard.cache.Load(blockID); found {
		shard.eviction.MoveToFront(element.(*list.Element))
		return
	}

	counter := maxEvictionRetriesIfFull
	for shard.currentSize+block.CurrentSize > shard.maxsize && counter > 0 {
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

func (bc *BlockCache) SearchBlock(blockID uint64, key string) (entity.Record, error) {
	block, found := bc.GetBlock(blockID)
	if found {
		return bc.searchInBlock(block, key)
	}
	return nil, errors.New("cache miss: block could not be found in the block cache")
}

func (bc *BlockCache) searchInBlock(block *Block, key string) (entity.Record, error) {
	record, err := block.GetRecord(key)
	if err != nil {
		return nil, err
	}
	return record, nil
}
