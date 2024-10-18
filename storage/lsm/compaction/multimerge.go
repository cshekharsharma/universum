package compaction

import (
	"container/heap"
	"universum/entity"
)

type MergeEntry struct {
	Key    string
	Record entity.Record
	Index  int
	Pos    int
}

type MinHeap []*MergeEntry

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].Key < h[j].Key
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*MergeEntry))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func MultiWayMerge(arrays [][]*entity.RecordKV) []*entity.RecordKV {
	h := &MinHeap{}
	heap.Init(h)

	for i, arr := range arrays {
		if len(arr) > 0 {
			heap.Push(h, &MergeEntry{
				Key:    arr[0].Key,
				Record: arr[0].Record,
				Index:  i,
				Pos:    0,
			})
		}
	}

	var result []*entity.RecordKV

	for h.Len() > 0 {
		smallest := heap.Pop(h).(*MergeEntry)
		result = append(result, &entity.RecordKV{Key: smallest.Key, Record: smallest.Record})

		if smallest.Pos+1 < len(arrays[smallest.Index]) {
			nextEntry := arrays[smallest.Index][smallest.Pos+1]
			heap.Push(h, &MergeEntry{
				Key:    nextEntry.Key,
				Record: nextEntry.Record,
				Index:  smallest.Index,
				Pos:    smallest.Pos + 1,
			})
		}
	}

	return result
}
