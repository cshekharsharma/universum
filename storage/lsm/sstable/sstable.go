package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"universum/config"
	"universum/dslib"
	"universum/entity"
	"universum/storage/lsm/memtable"
	"universum/utils"
)

type SSTable struct {
	filename  string
	fileptr   *os.File
	writeMode bool

	BloomFilter  *dslib.BloomFilter
	Index        map[string]int64
	CurrentBlock *Block

	RecordCount int64
	DataSize    int64
	Metadata    *Metadata
}

func NewSSTable(filename string, writeMode bool, maxRecords int64, falsePositiveRate float64) (*SSTable, error) {
	var file *os.File
	var err error

	datadir := config.Store.Storage.LSM.DataStorageDirectory
	sstFullPath := filepath.Clean(fmt.Sprintf("%s/%s", datadir, filename))

	if writeMode {
		file, err = os.Create(sstFullPath)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.Open(sstFullPath)
		if err != nil {
			return nil, err
		}
	}

	index := make(map[string]int64)

	bfSize, bfHashCount := dslib.OptimalBloomFilterSize(maxRecords, falsePositiveRate)
	bloomFilter := dslib.NewBloomFilter(bfSize, bfHashCount)

	metadata := &Metadata{
		Version:     1,
		Timestamp:   utils.GetCurrentEPochTime(),
		Compression: config.Store.Storage.LSM.BlockCompressionAlgo,
	}

	return &SSTable{
		filename:     filename,
		fileptr:      file,
		writeMode:    writeMode,
		Index:        index,
		BloomFilter:  bloomFilter,
		Metadata:     metadata,
		DataSize:     0,
		CurrentBlock: NewBlock(config.Store.Storage.LSM.WriteBlockSize),
		RecordCount:  0,
	}, nil
}

/////////////////////// Loader functions //////////////////////////

func (sst *SSTable) ReadBlock(blockOffset int64, blockSize int64) (*Block, error) {
	_, err := sst.fileptr.Seek(blockOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to block at offset %d: %v", blockOffset, err)
	}

	blockData := make([]byte, blockSize)
	_, err = sst.fileptr.Read(blockData)
	if err != nil {
		return nil, fmt.Errorf("failed to read block data: %v", err)
	}

	b := NewBlock(config.Store.Storage.LSM.WriteBlockSize)
	return b.DeserializeBlock(blockData, blockSize)
}

func (sst *SSTable) LoadIndex() error {
	if sst.Metadata.IndexOffset == 0 || sst.Metadata.IndexSize == 0 {
		return fmt.Errorf("index metadata is missing or invalid")
	}

	_, err := sst.fileptr.Seek(sst.Metadata.IndexOffset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to index offset: %v", err)
	}

	indexBytes := make([]byte, sst.Metadata.IndexSize)
	_, err = sst.fileptr.Read(indexBytes)
	if err != nil {
		return fmt.Errorf("failed to read index data: %v", err)
	}

	buf := bytes.NewBuffer(indexBytes)
	for {
		var keyLen int64
		err := binary.Read(buf, binary.BigEndian, &keyLen)
		if err != nil {
			break // End of index
		}

		key := make([]byte, keyLen)
		buf.Read(key)

		var offset int64
		err = binary.Read(buf, binary.BigEndian, &offset)
		if err != nil {
			return fmt.Errorf("failed to read block offset: %v", err)
		}

		sst.Index[string(key)] = offset
	}

	return nil
}

func (sst *SSTable) LoadBloomFilter() error {
	if sst.Metadata.BloomFilterOffset == 0 || sst.Metadata.BloomFilterSize == 0 {
		return fmt.Errorf("bloom filter metadata is missing or invalid")
	}

	_, err := sst.fileptr.Seek(sst.Metadata.BloomFilterOffset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to bloom filter offset: %v", err)
	}

	bloomFilterBytes := make([]byte, sst.Metadata.BloomFilterSize)
	_, err = sst.fileptr.Read(bloomFilterBytes)
	if err != nil {
		return fmt.Errorf("failed to read bloom filter data: %v", err)
	}

	bloomFilter := &dslib.BloomFilter{}
	err = bloomFilter.Deserialize(bloomFilterBytes)
	if err != nil {
		return fmt.Errorf("failed to deserialize bloom filter: %v", err)
	}

	sst.BloomFilter = bloomFilter
	return nil
}

func (sst *SSTable) LoadMetadata() error {
	fileInfo, err := sst.fileptr.Stat()
	if err != nil {
		return fmt.Errorf("failed to get SSTable file information: %v", err)
	}

	fileSize := fileInfo.Size()

	metadataSizeOffset := fileSize - int64(binary.Size(entity.Int64SizeInBytes))
	_, err = sst.fileptr.Seek(metadataSizeOffset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to metadata size position: %v", err)
	}

	var metadataSize int64
	err = binary.Read(sst.fileptr, binary.BigEndian, &metadataSize)
	if err != nil {
		return fmt.Errorf("failed to read metadata size: %v", err)
	}

	metadataOffset := metadataSizeOffset - metadataSize
	_, err = sst.fileptr.Seek(metadataOffset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to metadata block: %v", err)
	}

	metadataBytes := make([]byte, metadataSize)
	_, err = sst.fileptr.Read(metadataBytes)
	if err != nil {
		return fmt.Errorf("failed to read metadata block: %v", err)
	}

	metadata := &Metadata{}
	err = metadata.Deserialize(metadataBytes)
	if err != nil {
		return fmt.Errorf("failed to deserialize metadata: %v", err)
	}

	sst.Metadata = metadata
	return nil
}

/////////////////////// Writer functions /////////////////////////

func (sst *SSTable) FlushMemTableToSSTable(mem memtable.MemTable) error {
	records := mem.GetAllRecords()

	for key, value := range records {
		err := sst.CurrentBlock.AddRecord(key, value.ToMap(), sst.BloomFilter)
		if err == nil {
			continue // Successfully added record, continue
		}

		// if we are here, means the block is full
		err = sst.FlushBlock()
		if err != nil {
			return fmt.Errorf("failed to flush block to SSTable: %v", err)
		}

		// Add the record to the next block
		err = sst.CurrentBlock.AddRecord(key, value.ToMap(), sst.BloomFilter)
		if err != nil {
			return fmt.Errorf("failed to add record to new block after flushing: %v", err)
		}
	}

	// Flush the last block if it has any records
	if len(sst.CurrentBlock.records) > 0 {
		err := sst.FlushBlock()
		if err != nil {
			return fmt.Errorf("failed to flush last block to SSTable: %v", err)
		}
	}

	err := sst.FlushIndex()
	if err != nil {
		return fmt.Errorf("failed to write SSTable index: %v", err)
	}

	err = sst.FlushBloomFilter()
	if err != nil {
		return fmt.Errorf("failed to write bloom filter to SSTable %v", err)
	}

	err = sst.WriteMetadata()
	if err != nil {
		return fmt.Errorf("failed to write SSTable metadata: %v", err)
	}

	// forcefully sync the data to disk. This is important to make sure
	// that the OS doesnt keep that in OS write buffer for too long.
	// Otherwise, we might lose data in case of a system crash.
	err = sst.fileptr.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync SSTable file: %v", err)
	}

	return nil
}

func (sst *SSTable) FlushBlock() error {
	if len(sst.CurrentBlock.records) == 0 {
		return fmt.Errorf("no records to flush in the current block")
	}

	blockStartOffset, err := sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get current offset: %v", err)
	}

	serializedBlock, err := sst.CurrentBlock.SerializeBlock()
	if err != nil {
		return fmt.Errorf("failed to serialize block: %v", err)
	}

	_, err = sst.fileptr.Write(serializedBlock)
	if err != nil {
		return fmt.Errorf("failed to write block to SSTable file: %v", err)
	}

	flushedBlockSize := int64(len(serializedBlock))
	sst.RecordCount += int64(len(sst.CurrentBlock.records))
	sst.Metadata.DataSize += flushedBlockSize

	// Update the index with the block start offset and size by packing into single number
	// assumption is that neither the offset nor the block will ever cross 2G limit, hence
	// int32 number is safe to hold both of them respectively.
	sst.Index[sst.CurrentBlock.startKey] = utils.PackNumbers(int32(blockStartOffset), int32(flushedBlockSize))

	sst.CurrentBlock = NewBlock(sst.CurrentBlock.maxSize)
	return nil
}

func (sst *SSTable) FlushIndex() error {
	indexStartOffset, err := sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of SSTable file: %v", err)
	}

	buf := bytes.NewBuffer(make([]byte, sst.RecordCount))

	for key, offset := range sst.Index {
		keyBytes := []byte(key)
		keyLen := int64(len(keyBytes))

		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			return fmt.Errorf("failed to write key length: %v", err)
		}
		buf.Write(keyBytes)

		if err := binary.Write(buf, binary.BigEndian, offset); err != nil {
			return fmt.Errorf("failed to write block offset: %v", err)
		}
	}

	indexData := buf.Bytes()
	indexChecksum := crc32.ChecksumIEEE(indexData)

	if err := binary.Write(buf, binary.BigEndian, indexChecksum); err != nil {
		return fmt.Errorf("failed to append index checksum: %v", err)
	}

	_, err = sst.fileptr.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write SSTable index and checksum: %v", err)
	}

	sst.Metadata.IndexChecksum = indexChecksum
	sst.Metadata.IndexOffset = indexStartOffset
	sst.Metadata.IndexSize = int64(len(indexData) + binary.Size(indexChecksum))

	sst.Metadata.DataSize += sst.Metadata.IndexSize
	return nil
}

func (sst *SSTable) FlushBloomFilter() error {
	if sst.BloomFilter == nil {
		return fmt.Errorf("bloom filter is nil")
	}

	bloomFilterOffset, err := sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of the file: %v", err)
	}

	serializedBloomFilter, err := sst.BloomFilter.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize bloom filter: %v", err)
	}

	bloomFilterSize := int64(len(serializedBloomFilter))
	err = binary.Write(sst.fileptr, binary.BigEndian, bloomFilterSize)
	if err != nil {
		return fmt.Errorf("failed to write bloom filter size: %v", err)
	}

	_, err = sst.fileptr.Write(serializedBloomFilter)
	if err != nil {
		return fmt.Errorf("failed to write bloom filter to SSTable file: %v", err)
	}

	sst.Metadata.BloomFilterOffset = bloomFilterOffset
	sst.Metadata.BloomFilterSize = (bloomFilterSize + int64(binary.Size(bloomFilterSize)))

	sst.Metadata.DataSize += sst.Metadata.BloomFilterSize
	return nil
}

func (sst *SSTable) WriteMetadata() error {
	sst.Metadata.NumRecords = sst.RecordCount
	sst.Metadata.Timestamp = utils.GetCurrentEPochTime()

	metadataBytes, err := sst.Metadata.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %v", err)
	}

	_, err = sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get current offset: %v", err)
	}

	_, err = sst.fileptr.Write(metadataBytes)
	if err != nil {
		return fmt.Errorf("failed to write metadata to SSTable: %v", err)
	}

	metadataSize := int64(len(metadataBytes))
	err = binary.Write(sst.fileptr, binary.BigEndian, metadataSize)
	if err != nil {
		return fmt.Errorf("failed to write metadata size to SSTable: %v", err)
	}

	sst.DataSize = sst.Metadata.DataSize
	sst.DataSize += int64(len(metadataBytes) + binary.Size(metadataSize))

	return nil
}
