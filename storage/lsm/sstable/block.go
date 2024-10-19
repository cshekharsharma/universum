package sstable

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"io"
	"sort"
	"sync"
	"universum/dslib"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
)

type Block struct {
	Id          uint64
	Records     []*entity.SerializedRecordKV
	Index       sync.Map
	CurrentSize int64
	MaxSize     int64
	FirstKey    string
	LastKey     string
	Data        []byte
	Checksum    uint32
}

func NewBlock(maxSize int64) *Block {
	return &Block{
		Id:          0,
		Records:     make([]*entity.SerializedRecordKV, 0),
		Index:       sync.Map{},
		CurrentSize: 0,
		MaxSize:     maxSize,
	}
}

func (b *Block) GetID() uint64 {
	return b.Id
}

func (b *Block) SetID(id uint64) {
	b.Id = id
}

func (b *Block) GetKeyInfoFromIndex(key string) (bool, int64, error) {
	if recordOffset, found := b.Index.Load(key); found {
		return true, recordOffset.(int64), nil
	}

	return false, 0, nil
}

func (b *Block) GetRecord(key string) (entity.Record, error) {
	keyExists, recordOffset, err := b.GetKeyInfoFromIndex(key)
	if err != nil || !keyExists {
		return nil, nil
	}

	_, encodedRecord, err := b.ReadRecordAtOffset(recordOffset)
	if err != nil {
		return nil, fmt.Errorf("error in reading record for key '%s': %v", key, err)
	}

	record, err := resp3.GetScalarRecordFromResp(string(encodedRecord))
	if err != nil {
		return nil, fmt.Errorf("record for '%s' found in invalid format", key)
	}

	return record, nil
}

func (b *Block) AddRecord(key string, value map[string]interface{}, bloom *dslib.BloomFilter) error {
	serialisedValue, err := resp3.Encode(value)
	if err != nil {
		logger.Get().Warn("failed to encode record value for key=%s: %v", key, err)
		return nil // ignore the record and move on.
	}

	valueSize := len(serialisedValue)
	recordSize := int64(len(key)+valueSize) + 2*entity.Int64SizeInBytes

	if b.CurrentSize+recordSize > b.MaxSize {
		return fmt.Errorf("Block is full, cannot add more records [size=%d, maxSize=%d]", b.CurrentSize, b.MaxSize)
	}

	b.Records = append(b.Records, &entity.SerializedRecordKV{
		Key:    []byte(key),
		Record: []byte(serialisedValue),
	})

	if len(b.Records) == 1 {
		b.FirstKey = key
	}
	b.LastKey = key

	bloom.Add(key)
	b.CurrentSize += recordSize
	return nil
}

func (b *Block) SerializeBlock() ([]byte, error) {
	var currentOffset int64 = 0
	buf := bytes.NewBuffer(make([]byte, 0, b.MaxSize)) // preset approx size

	for i := 0; i < len(b.Records); i++ {
		keyBytes := b.Records[i].Key
		valueBytes := b.Records[i].Record

		keyLen := int64(len(keyBytes))
		valueLen := int64(len(valueBytes))

		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			return nil, err
		}
		buf.Write(keyBytes)

		if err := binary.Write(buf, binary.BigEndian, valueLen); err != nil {
			return nil, err
		}
		buf.Write(valueBytes)

		b.Index.Store(string(b.Records[i].Key), currentOffset)
		currentOffset += int64(entity.Int64SizeInBytes + keyLen + entity.Int64SizeInBytes + valueLen)
	}

	// Serialize the block index at the end of the block
	// Each index entry is [key length + key + offset]
	var indexsize int64 = 0
	var readErr error
	b.Index.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		offset := value.(int64)

		keyBytes := []byte(keyStr)
		keyLen := int64(len(keyBytes))

		if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
			readErr = err
			return false
		}
		buf.Write(keyBytes)

		if err := binary.Write(buf, binary.BigEndian, offset); err != nil {
			readErr = err
			return false
		}

		indexsize += int64(binary.Size(keyLen)) + keyLen + int64(binary.Size(offset))
		return true
	})

	if readErr != nil {
		return nil, readErr
	}

	if err := binary.Write(buf, binary.BigEndian, indexsize); err != nil {
		return nil, err
	}

	b.Data = buf.Bytes()
	b.Checksum = crc32.ChecksumIEEE(b.Data)

	if err := binary.Write(buf, binary.BigEndian, b.Checksum); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Block) ReadRecordAtOffset(offset int64) ([]byte, []byte, error) {
	keyLenSize := int64(entity.Int64SizeInBytes)

	if offset+keyLenSize > int64(len(b.Data)) {
		return nil, nil, fmt.Errorf("offset exceeds block size")
	}
	keyLen := binary.BigEndian.Uint64(b.Data[offset : offset+keyLenSize])

	startKey := offset + keyLenSize
	endKey := startKey + int64(keyLen)
	if endKey > int64(len(b.Data)) {
		return nil, nil, fmt.Errorf("key exceeds block size")
	}
	key := b.Data[startKey:endKey]

	endKeyLenSize := endKey + keyLenSize
	if endKeyLenSize > int64(len(b.Data)) {
		return nil, nil, fmt.Errorf("value length exceeds block size")
	}
	valueLen := binary.BigEndian.Uint64(b.Data[endKey:endKeyLenSize])

	startValue := endKeyLenSize
	endValue := startValue + int64(valueLen)
	if endValue > int64(len(b.Data)) {
		return nil, nil, fmt.Errorf("value exceeds block size")
	}

	value := b.Data[startValue:endValue]
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

		b.Index.Store(string(keyBytes), recordOffset)
	}

	b.Data = blockData[:blockChecksumOffset]
	b.Checksum = storedChecksum

	return b, nil
}

func (b *Block) GetAllRecords() ([]*entity.RecordKV, error) {
	buf := bytes.NewReader(b.Data)
	var readErr error

	recordList := make([]*entity.RecordKV, 0)

	b.Index.Range(func(_, value interface{}) bool {
		offset := value.(int64)
		buf.Seek(offset, io.SeekStart)

		keyBytes, value, err := b.ReadRecordAtOffset(offset)
		if err != nil {
			readErr = err
			return false
		}

		decodedRecord, err := resp3.Decode(bufio.NewReader(bytes.NewReader(value.([]byte))))
		if err != nil {
			return true // ignore the faulty record, and move on.
		}

		_, record := (&entity.ScalarRecord{}).FromMap(decodedRecord.(map[string]interface{}))

		recordList = append(recordList, &entity.RecordKV{
			Key:    string(keyBytes),
			Record: record,
		})

		return true
	})

	if readErr != nil {
		return nil, readErr
	}

	// @TODO, ideally shouldn not have been required, and is an overhead.
	// better to avoid iterating over syncMap and read in sequential order from
	// serialised byte array maintained within block.
	sort.Slice(recordList, func(i, j int) bool {
		return recordList[i].Key < recordList[j].Key
	})

	return recordList, nil
}

func (b *Block) ValidateBlock() bool {
	return crc32.ChecksumIEEE(b.Data) == b.Checksum
}

func (b *Block) RemainingSpace() int64 {
	return b.MaxSize - b.CurrentSize
}

func GenerateBlockID(firstKey string, lastKey string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(fmt.Sprintf("f:%s-l:%s", firstKey, lastKey)))
	return hash.Sum64()
}
