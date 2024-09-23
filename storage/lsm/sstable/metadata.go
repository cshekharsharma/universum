package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

type Metadata struct {
	Version           int64  // Version of the SSTable format
	NumRecords        int64  // Number of key-value pairs in the SSTable
	IndexOffset       int64  // Offset in the file where the index block starts
	DataSize          int64  // Total size of the data block
	IndexSize         int64  // Size of the index block
	BloomFilterOffset int64  // Offset for the bloom filter block (if used)
	BloomFilterSize   int64  // Size of the bloom filter block
	Timestamp         int64  // Timestamp when this SSTable was created
	Compression       string // Compression algorithm used (if any)
	Checksum          uint32 // Optional checksum for data validation
	IndexChecksum     uint32 // Optional checksum for index block validation
}

func (m *Metadata) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	val := reflect.ValueOf(m).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i).Interface()
		if err := binary.Write(buf, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to serialize field %s: %v", val.Type().Field(i).Name, err)
		}
	}
	return buf.Bytes(), nil
}

func (m *Metadata) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)
	val := reflect.ValueOf(m).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i).Addr().Interface()
		if err := binary.Read(buf, binary.BigEndian, field); err != nil {
			return fmt.Errorf("failed to deserialize field %s: %v", val.Type().Field(i).Name, err)
		}
	}

	return nil
}
