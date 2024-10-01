package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
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

func NewBlock(blockSize int64) *Block {
	return &Block{
		records:     make(map[string]string),
		index:       make(map[string]int64),
		currentSize: 0,
		maxSize:     blockSize,
	}
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
	recordSize := int64(len(key) + int(valueSize) + 2*entity.Int64SizeInBytes)

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

	keyLen := binary.BigEndian.Uint64(b.data[offset : offset+keyLenSize])
	startKey := offset + keyLenSize
	endKey := startKey + int64(keyLen)
	key := b.data[startKey:endKey]

	valueLen := binary.BigEndian.Uint64(b.data[endKey : endKey+keyLenSize])
	startValue := endKey + keyLenSize
	endValue := startValue + int64(valueLen)
	value := b.data[startValue:endValue]

	return key, value, nil
}

func (b *Block) ValidateBlock() bool {
	return crc32.ChecksumIEEE(b.data) == b.checksum
}

func (b *Block) RemainingSpace() int64 {
	return b.maxSize - b.currentSize
}
