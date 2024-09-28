package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"universum/config"
	"universum/dslib"
	"universum/entity"
	"universum/resp3"
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
		filename:     filename,
		fileptr:      file,
		writeMode:    writeMode,
		Index:        index,
		BloomFilter:  bloomFilter,
		Metadata:     metadata,
		DataSize:     0,
		CurrentBlock: NewBlock(config.Store.Storage.LSM.WriteBlockSize),
	}, nil
}

/////////////////////// Loader functions //////////////////////////

func (sst *SSTable) ReadBlock(blockOffset int64) (*Block, error) {
	_, err := sst.fileptr.Seek(blockOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to block at offset %d: %v", blockOffset, err)
	}

	var blockSize int64
	err = binary.Read(sst.fileptr, binary.BigEndian, &blockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read block size: %v", err)
	}

	blockData := make([]byte, blockSize)
	_, err = sst.fileptr.Read(blockData)
	if err != nil {
		return nil, fmt.Errorf("failed to read block data: %v", err)
	}

	var storedChecksum uint32
	err = binary.Read(sst.fileptr, binary.BigEndian, &storedChecksum)
	if err != nil {
		return nil, fmt.Errorf("failed to read block checksum: %v", err)
	}

	calculatedChecksum := crc32.ChecksumIEEE(blockData)
	if storedChecksum != calculatedChecksum {
		return nil, fmt.Errorf("block checksum validation failed")
	}

	block := &Block{
		records:  make(map[string]interface{}),
		index:    make(map[string]int64),
		data:     blockData,
		checksum: storedChecksum,
	}

	buf := bytes.NewReader(blockData)
	var currentOffset int64 = 0

	for currentOffset < blockSize {
		// Read key length
		var keyLen int64
		err := binary.Read(buf, binary.BigEndian, &keyLen)
		if err != nil {
			return nil, fmt.Errorf("failed to read key length: %v", err)
		}

		// Read the key
		key := make([]byte, keyLen)
		_, err = buf.Read(key)
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %v", err)
		}

		// Read value length
		var valueLen int64
		err = binary.Read(buf, binary.BigEndian, &valueLen)
		if err != nil {
			return nil, fmt.Errorf("failed to read value length: %v", err)
		}

		// Read the value
		value := make([]byte, valueLen)
		_, err = buf.Read(value)
		if err != nil {
			return nil, fmt.Errorf("failed to read value: %v", err)
		}

		decodedValue, err := resp3.Decode(bufio.NewReader(bytes.NewReader(value)))
		if err != nil {
			return nil, fmt.Errorf("failed to decode value: %v", err)
		}

		block.records[string(key)] = decodedValue
		block.index[string(key)] = currentOffset

		currentOffset += int64(binary.Size(keyLen)) + keyLen + int64(binary.Size(valueLen)) + valueLen
	}

	return block, nil
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
		err := sst.CurrentBlock.AddRecord(key, value, sst.BloomFilter)
		if err == nil {
			continue // Successfully added record, continue
		}

		// if we are here, means the block is full
		err = sst.FlushBlock()
		if err != nil {
			return fmt.Errorf("failed to flush block to SSTable: %v", err)
		}

		// Add the record to the next block
		err = sst.CurrentBlock.AddRecord(key, value, sst.BloomFilter)
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

	// update SST index to keep start offset of block for each key
	sst.Index[sst.CurrentBlock.startKey] = blockStartOffset

	sst.RecordCount += int64(len(sst.CurrentBlock.records))
	sst.DataSize += int64(len(serializedBlock))

	sst.CurrentBlock = NewBlock(sst.CurrentBlock.maxSize)
	return nil
}

func (sst *SSTable) FlushIndex() error {
	_, err := sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of SSTable file: %v", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0))

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

	indexdata := buf.Bytes()
	_, err = sst.fileptr.Write(indexdata)
	if err != nil {
		return fmt.Errorf("failed to write SSTable index: %v", err)
	}

	sst.Metadata.IndexChecksum = crc32.ChecksumIEEE(indexdata)
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

	_, err = sst.fileptr.Write(serializedBloomFilter)
	if err != nil {
		return fmt.Errorf("failed to write bloom filter to SSTable file: %v", err)
	}

	sst.Metadata.BloomFilterOffset = bloomFilterOffset
	sst.Metadata.BloomFilterSize = int64(len(serializedBloomFilter))

	return nil
}

func (sst *SSTable) WriteMetadata() error {
	sst.Metadata.NumRecords = sst.RecordCount
	sst.Metadata.DataSize = sst.DataSize
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

	return nil
}
