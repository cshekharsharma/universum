package compaction

import "universum/storage/lsm/sstable"

type SSTReplacement struct {
	Obsoletes  []*sstable.SSTable
	Substitute *sstable.SSTable
}

var SSTReplacementChan chan *SSTReplacement
