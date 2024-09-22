package sstable

import (
	"os"
	"universum/dslib"
	"universum/utils"
)

type SSTable struct {
	filename  string
	fileptr   *os.File
	writeMode bool

	bloomFilter *dslib.BloomFilter
	index       map[string]int64
	blocks      []*Block
	currBlock   *Block
	blockIndex  map[string]int64

	recordCount int64
	dataSize    int64
	metadata    *Metadata
}

func NewSSTable(filename string, writeMode bool, maxRecords uint64, falsePositiveRate float64) (*SSTable, error) {
	var file *os.File
	var err error

	if writeMode {
		file, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}

	index := make(map[string]int64)

	bfSize, bfHashCount := dslib.OptimalBloomFilterSize(maxRecords, falsePositiveRate)
	bloomFilter := dslib.NewBloomFilter(bfSize, bfHashCount)

	metadata := &Metadata{
		Version:   1,
		Timestamp: utils.GetCurrentEPochTime(),
	}

	return &SSTable{
		filename:    filename,
		fileptr:     file,
		blocks:      []*Block{},
		blockIndex:  index,
		bloomFilter: bloomFilter,
		metadata:    metadata,
		dataSize:    0,
		currBlock:   NewBlock(),
		writeMode:   writeMode,
	}, nil
}

func (sst *SSTable) Get(key string) (interface{}, error) {
	// Implementation here
	return nil, nil
}

func (sst *SSTable) Exists(key string) (bool, error) {
	// Implementation here
	return false, nil
}

func (sst *SSTable) LoadIndex() error {
	// Implementation here
	return nil
}

func (sst *SSTable) LoadBloomFilter() error {
	// Implementation here
	return nil
}

func ReadMetadata() (*Metadata, error) {
	return nil, nil
}

func (sst *SSTable) FlushMemtableToFile(map[string]interface{}) error {
	// Implementation here
	return nil
}

func (sst *SSTable) MergeSSTables([]*SSTable) (*SSTable, error) {
	// Implementation here
	return nil, nil
}

func (sst *SSTable) ValidateSSTable() error {
	// Implementation here
	return nil
}

func (sst *SSTable) flushBlock() error {
	// Serialize block and write to file
	offset, err := sst.fileptr.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	sst.blocks = append(sst.blocks, sst.currBlock)
	sst.blockIndex[sst.currBlock.StartKey] = offset

	sst.currBlock = NewBlock()
	return nil
}
