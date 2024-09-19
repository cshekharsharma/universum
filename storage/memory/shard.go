package memory

import "sync"

type Shard struct {
	id   int64
	data *sync.Map
}

func NewShard(id int64) *Shard {
	return &Shard{
		id: id,
	}
}

func (s *Shard) GetId() int64 {
	return s.id
}

func (s *Shard) GetData() *sync.Map {
	return s.data
}
