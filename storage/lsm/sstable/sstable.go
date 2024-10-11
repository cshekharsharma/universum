package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"universum/compression"
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
	Index        []*sstIndexEntry
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

	writeBufferSize := config.Store.Storage.LSM.WriteBufferSize
	writeBlockSize := config.Store.Storage.LSM.WriteBlockSize

	// initialise with approximate number of blocks in the sst.
	index := make([]*sstIndexEntry, 0, writeBufferSize/writeBlockSize)

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

func (sst *SSTable) LoadSSTableFromDisk() error {
	err := sst.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata for SSTable %s: %v", sst.filename, err)
	}

	err = sst.LoadBloomFilter()
	if err != nil {
		return fmt.Errorf("failed to load Bloom filter for SSTable %s: %v", sst.filename, err)
	}

	err = sst.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index for SSTable %s: %v", sst.filename, err)
	}

	sst.DataSize = sst.Metadata.DataSize
	sst.RecordCount = sst.Metadata.NumRecords
	sst.CurrentBlock = NewBlock(config.Store.Storage.LSM.WriteBlockSize)

	return nil
}

func (sst *SSTable) LoadBlock(blockOffset int64, blockSize int64) (*Block, error) {
	_, err := sst.fileptr.Seek(blockOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to block at offset %d: %v", blockOffset, err)
	}

	blockData := make([]byte, blockSize)
	_, err = sst.fileptr.Read(blockData)
	if err != nil {
		return nil, fmt.Errorf("failed to read block data: %v", err)
	}

	compressor := compression.GetCompressor(&compression.Options{
		CompressionAlgo: compression.CompressionAlgo(sst.Metadata.Compression),
	})

	blockData, err = compressor.Decompress(blockData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress block data: %v", err)
	}

	b := NewBlock(config.Store.Storage.LSM.WriteBlockSize)
	return b.DeserializeBlock(blockData, int64(len(blockData)))
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

	calculatedChecksum := crc32.ChecksumIEEE(indexBytes)
	if calculatedChecksum != sst.Metadata.IndexChecksum {
		return fmt.Errorf("SST index checksum mismatch, possibly corrupt file")
	}

	offset := 0
	for offset < len(indexBytes) {
		var indexEntryLen int64
		err := binary.Read(bytes.NewReader(indexBytes[offset:]), binary.BigEndian, &indexEntryLen)

		if err != nil {
			if err == io.EOF {
				break // End of index
			}
			return fmt.Errorf("failed to read index entry length: %v", err)
		}

		offset += int(binary.Size(indexEntryLen))

		if indexEntryLen < 0 || offset+int(indexEntryLen) > len(indexBytes) {
			return fmt.Errorf("invalid index entry length: %d", indexEntryLen)
		}

		indexEntryStr := indexBytes[offset : offset+int(indexEntryLen)]
		offset += int(indexEntryLen)

		indexEntry, err := resp3.Decode(bufio.NewReader(bytes.NewReader(indexEntryStr)))
		if err != nil {
			return fmt.Errorf("failed to decode SST index entry read from disk: %v", err)
		}

		if asMap, ok := indexEntry.(map[string]interface{}); ok {
			sst.Index = append(sst.Index, &sstIndexEntry{
				f: asMap["f"].(string),
				l: asMap["l"].(string),
				o: asMap["o"].(int64),
			})
		}
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

	metadataSizeLength := int64(binary.Size(int64(0)))
	metadataSizeOffset := fileSize - metadataSizeLength
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

func (sst *SSTable) FindRecord(key string) (bool, entity.Record, error) {
	if key == "" {
		return false, nil, errors.New("empty key provided")
	}

	if sst.Metadata.FirstKey > key || sst.Metadata.LastKey < key {
		return false, nil, nil
	}

	if sst.BloomFilter != nil && !sst.BloomFilter.Exists(key) {
		return false, nil, nil
	}

	if indexEntry, err := sst.FindBlockForKey(key, sst.Index); err == nil {
		blockOffset, blockSize := utils.UnpackNumbers(indexEntry.GetOffset())
		block, err := sst.LoadBlock(int64(blockOffset), int64(blockSize))

		if err != nil {
			return false, nil, nil
		}

		record, err := block.GetRecord(key)
		if record == nil && err == nil {
			return false, nil, nil
		}

		if err != nil {
			return false, nil, err
		}

		return true, record, nil
	}

	return false, nil, nil
}

func (sst *SSTable) FindBlockForKey(key string, index []*sstIndexEntry) (*sstIndexEntry, error) {
	idx := sort.Search(len(index), func(i int) bool {
		return sst.Index[i].GetFirstKey() > key
	})

	if idx == 0 || idx >= len(sst.Index) {
		return nil, fmt.Errorf("key not found in index")
	}

	entry := sst.Index[idx-1]
	if key >= entry.GetFirstKey() && key <= entry.GetLastKey() {
		return entry, nil
	}

	return nil, fmt.Errorf("key not found in any block")
}

/////////////////////// Writer functions /////////////////////////

func (sst *SSTable) FlushMemTableToSSTable(mem memtable.MemTable) error {
	recordList := mem.GetAll()

	for i := 0; i < len(recordList); i++ {
		key := recordList[i].Key
		err := sst.CurrentBlock.AddRecord(key, recordList[i].Record.ToMap(), sst.BloomFilter)
		if err == nil {
			continue // Successfully added record, continue
		}

		// if we are here, means the block is full
		err = sst.FlushBlock()
		if err != nil {
			return fmt.Errorf("failed to flush block to SSTable: %v", err)
		}

		// Add the record to the next block
		err = sst.CurrentBlock.AddRecord(key, recordList[i].Record.ToMap(), sst.BloomFilter)
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

	err = sst.FlushMetadata()
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

	compressor := compression.GetCompressor(&compression.Options{
		CompressionAlgo: compression.CompressionAlgo(sst.Metadata.Compression),
		Writer:          sst.fileptr,
		AutoCloseWriter: true,
	})

	serializedBlock, err = compressor.Compress(serializedBlock)
	if err != nil {
		return fmt.Errorf("failed to compress block: %v", err)
	}

	_, err = sst.fileptr.Write(serializedBlock)
	if err != nil {
		return fmt.Errorf("failed to write block to SSTable file: %v", err)
	}

	flushedBlockSize := int64(len(serializedBlock))
	sst.RecordCount += int64(len(sst.CurrentBlock.records))
	sst.Metadata.DataSize += flushedBlockSize
	sst.Metadata.NumRecords += int64(len(sst.CurrentBlock.records))

	// Update the index with the block start offset and size by packing into single number
	// assumption is that neither the offset nor the block will ever cross 2G limit, hence
	// int32 number is safe enough to hold both of them respectively.
	// @FIXME index should hold both start and end keys and respective offset info
	blockOffsetAndLength := utils.PackNumbers(int32(blockStartOffset), int32(flushedBlockSize))
	sst.Index = append(sst.Index, &sstIndexEntry{
		f: sst.CurrentBlock.firstKey,
		l: sst.CurrentBlock.lastKey,
		o: blockOffsetAndLength,
	})

	if sst.Metadata.FirstKey == "" {
		sst.Metadata.FirstKey = sst.CurrentBlock.firstKey
	}
	sst.Metadata.LastKey = sst.CurrentBlock.lastKey

	sst.CurrentBlock = NewBlock(sst.CurrentBlock.maxSize)
	return nil
}

func (sst *SSTable) FlushIndex() error {
	indexStartOffset, err := sst.fileptr.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of SSTable file: %v", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, sst.RecordCount*16))

	for _, indexEntry := range sst.Index {
		encodedIdxEntry, err := resp3.Encode(indexEntry.ToMap())
		if err != nil {
			return fmt.Errorf("failed to encode index object: %v", err)
		}

		asBytes := []byte(encodedIdxEntry)
		byteLength := int64(len(asBytes))

		if err := binary.Write(buf, binary.BigEndian, byteLength); err != nil {
			return fmt.Errorf("failed to write key length: %v", err)
		}
		buf.Write(asBytes)
	}

	indexData := buf.Bytes()
	indexChecksum := crc32.ChecksumIEEE(indexData)

	_, err = sst.fileptr.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write SSTable index and checksum: %v", err)
	}

	sst.Metadata.IndexChecksum = indexChecksum
	sst.Metadata.IndexOffset = indexStartOffset
	sst.Metadata.IndexSize = int64(len(indexData))

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

	_, err = sst.fileptr.Write(serializedBloomFilter)
	if err != nil {
		return fmt.Errorf("failed to write bloom filter to SSTable file: %v", err)
	}

	sst.Metadata.BloomFilterOffset = bloomFilterOffset
	sst.Metadata.BloomFilterSize = bloomFilterSize

	sst.Metadata.DataSize += sst.Metadata.BloomFilterSize
	return nil
}

func (sst *SSTable) FlushMetadata() error {
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
