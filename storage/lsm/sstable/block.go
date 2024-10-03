package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"universum/dslib"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
)

type Block struct {
	records     map[string]string
	index       map[string]int64
	currentSize int64
	maxSize     int64
	startKey    string
	endKey      string
	data        []byte
	checksum    uint32
}

func NewBlock(maxSize int64) *Block {
	return &Block{
		records:     make(map[string]string),
		index:       make(map[string]int64),
		currentSize: 0,
		maxSize:     maxSize,
	}
}

func (b *Block) GetKeyInfoFromIndex(key string) (bool, int64, error) {
	if b.index == nil {
		return false, 0, errors.New("block index is not initialised")
	}

	if recordOffset, found := b.index[key]; found {
		return true, recordOffset, nil
	}

	return false, 0, nil
}

func (b *Block) GetRecord(key string) (entity.Record, error) {
	serialisedValue, ok := b.records[key]
	if !ok {
		return nil, errors.New("record not found")
	}

	value, err := resp3.Decode(bufio.NewReader(bytes.NewReader([]byte(serialisedValue))))
	if err != nil {
		return nil, fmt.Errorf("record found, but unable to decode: err=%v", err)
	}

	if _, ok := value.(*map[interface{}]interface{}); !ok {
		return nil, errors.New("record found, but not in the correct format")
	}

	record := value.(map[interface{}]interface{})
	return &entity.ScalarRecord{
		Value:  record["Value"],
		Type:   record["Type"].(uint8),
		LAT:    record["LAT"].(int64),
		Expiry: record["Expiry"].(int64),
	}, nil
}

func (b *Block) AddRecord(key string, value map[string]interface{}, bloom *dslib.BloomFilter) error {
	serialisedValue, err := resp3.Encode(value)
	if err != nil {
		logger.Get().Warn("failed to encode record value: %v", err)
		return nil // ignore the record and move on.
	}

	valueSize := len(serialisedValue)
	recordSize := int64(len(key)+valueSize) + 2*entity.Int64SizeInBytes

	if b.currentSize+recordSize > b.maxSize {
		return fmt.Errorf("Block is full, cannot add more records [size=%d, maxSize=%d]", b.currentSize, b.maxSize)
	}

	b.records[key] = serialisedValue

	if len(b.records) == 1 {
		b.startKey = key
	}
	b.endKey = key

	bloom.Add(key)
	b.currentSize += recordSize
	return nil
}

func (b *Block) SerializeBlock() ([]byte, error) {
	var currentOffset int64 = 0
	buf := bytes.NewBuffer(make([]byte, 0, b.maxSize)) // preset approx size

	for key, value := range b.records {
		keyBytes := []byte(key)
		valueBytes := []byte(value)

		keyLen := int64(len(keyBytes))
		valueLen := int64(len(valueBytes))

		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			return nil, err
		}
		buf.Write(keyBytes)

		if err := binary.Write(buf, binary.BigEndian, valueLen); err != nil {
			return nil, err
		}
		buf.Write([]byte(valueBytes))

		b.index[key] = currentOffset
		currentOffset += int64(entity.Int64SizeInBytes + keyLen + entity.Int64SizeInBytes + valueLen)
	}

	// Serialize the block index at the end of the block
	// Each index entry is [key length + key + offset]
	var indexsize int64 = 0
	for key, offset := range b.index {
		keyBytes := []byte(key)
		keyLen := int64(len(keyBytes))

		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			return nil, err
		}
		buf.Write(keyBytes)

		if err := binary.Write(buf, binary.BigEndian, offset); err != nil {
			return nil, err
		}

		indexsize += int64(binary.Size(keyLen)) + keyLen + int64(binary.Size(offset))
	}

	if err := binary.Write(buf, binary.BigEndian, indexsize); err != nil {
		return nil, err
	}

	b.data = buf.Bytes()
	b.checksum = crc32.ChecksumIEEE(b.data)

	if err := binary.Write(buf, binary.BigEndian, b.checksum); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Block) ReadRecordAtOffset(offset int64) ([]byte, []byte, error) {
	keyLenSize := int64(entity.Int64SizeInBytes)

	if offset+keyLenSize > int64(len(b.data)) {
		return nil, nil, fmt.Errorf("offset exceeds block size")
	}
	keyLen := binary.BigEndian.Uint64(b.data[offset : offset+keyLenSize])

	startKey := offset + keyLenSize
	endKey := startKey + int64(keyLen)
	if endKey > int64(len(b.data)) {
		return nil, nil, fmt.Errorf("key exceeds block size")
	}
	key := b.data[startKey:endKey]

	endKeyLenSize := endKey + keyLenSize
	if endKeyLenSize > int64(len(b.data)) {
		return nil, nil, fmt.Errorf("value length exceeds block size")
	}
	valueLen := binary.BigEndian.Uint64(b.data[endKey:endKeyLenSize])

	startValue := endKeyLenSize
	endValue := startValue + int64(valueLen)
	if endValue > int64(len(b.data)) {
		return nil, nil, fmt.Errorf("value exceeds block size")
	}

	value := b.data[startValue:endValue]
	return key, value, nil
}

func (b *Block) DeserializeBlock(blockData []byte, blockSize int64) (*Block, error) {
	blockChecksumOffset := blockSize - int64(entity.Int32SizeInBytes)
	storedChecksum := binary.BigEndian.Uint32(blockData[blockChecksumOffset:])
	calculatedChecksum := crc32.ChecksumIEEE(blockData[:blockChecksumOffset])

	if storedChecksum != calculatedChecksum {
		return nil, fmt.Errorf("block checksum validation failed")
	}

	indexSizeOffset := len(blockData) - entity.Int32SizeInBytes - entity.Int64SizeInBytes
	indexSize := int64(binary.BigEndian.Uint64(blockData[indexSizeOffset:]))

	indexOffset := indexSizeOffset - int(indexSize)
	indexData := blockData[indexOffset : indexOffset+int(indexSize)]

	blockIndex := make(map[string]int64)
	buf := bytes.NewReader(indexData)

	for {
		var keyLen int64
		err := binary.Read(buf, binary.BigEndian, &keyLen)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read key length from index: %v", err)
		}

		keyBytes := make([]byte, keyLen)
		_, err = buf.Read(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read key from index: %v", err)
		}

		var recordOffset int64
		err = binary.Read(buf, binary.BigEndian, &recordOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read record offset from index: %v", err)
		}

		blockIndex[string(keyBytes)] = recordOffset
	}

	b.index = blockIndex
	b.data = blockData[:blockChecksumOffset]
	b.checksum = storedChecksum

	return b, nil
}

func (b *Block) PopulateRecordsInBlock() (*Block, error) {
	buf := bytes.NewReader(b.data)

	for _, offset := range b.index {
		buf.Seek(offset, io.SeekStart)

		keyBytes, value, err := b.ReadRecordAtOffset(offset)
		if err != nil {
			return nil, err
		}

		b.records[string(keyBytes)] = string(value)
	}

	return b, nil
}

func (b *Block) ValidateBlock() bool {
	return crc32.ChecksumIEEE(b.data) == b.checksum
}

func (b *Block) RemainingSpace() int64 {
	return b.maxSize - b.currentSize
}
