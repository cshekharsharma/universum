package memory

import (
	"sync"
	"testing"
)

func TestNewShard(t *testing.T) {
	shardID := int64(1)
	shard := NewShard(shardID)

	if shard == nil {
		t.Fatal("Expected NewShard to return a non-nil Shard")
	}

	if shard.id != shardID {
		t.Errorf("NewShard() shard.id = %v, want %v", shard.id, shardID)
	}

	if shard.data != nil {
		t.Errorf("NewShard() expected shard.data to be nil, got %v", shard.data)
	}
}

func TestShard_GetId(t *testing.T) {
	shardID := int64(2)
	shard := &Shard{id: shardID}

	gotID := shard.GetId()
	if gotID != shardID {
		t.Errorf("Shard.GetId() = %v, want %v", gotID, shardID)
	}
}

func TestShard_GetData(t *testing.T) {
	shard := &Shard{
		data: &sync.Map{},
	}

	data := shard.GetData()
	if data == nil {
		t.Error("Shard.GetData() expected non-nil sync.Map, got nil")
	}
}
