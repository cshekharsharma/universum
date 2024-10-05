package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Metadata represents the metadata information for an SSTable.
type Metadata struct {
	Version           int64  // Version of the SSTable format
	NumRecords        int64  // Number of key-value pairs in the SSTable
	DataSize          int64  // Total size of the data block
	IndexOffset       int64  // Offset in the file where the index block starts
	IndexSize         int64  // Size of the index block
	IndexChecksum     uint32 // Optional checksum for index block validation
	BloomFilterOffset int64  // Offset for the bloom filter block (if used)
	BloomFilterSize   int64  // Size of the bloom filter block
	Timestamp         int64  // Timestamp when this SSTable was created
	Compression       string // Compression algorithm used (if any)
}

// Serialize converts the Metadata struct into a byte slice using binary encoding.
// It handles fixed-size fields and variable-length string fields.
func (m *Metadata) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	fixedFields := []interface{}{
		m.Version,
		m.NumRecords,
		m.DataSize,
		m.IndexOffset,
		m.IndexSize,
		m.IndexChecksum,
		m.BloomFilterOffset,
		m.BloomFilterSize,
		m.Timestamp,
	}

	for _, field := range fixedFields {
		if err := binary.Write(buf, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to serialize: %v", err)
		}
	}

	str := m.Compression
	strLen := int32(len(str))
	if err := binary.Write(buf, binary.BigEndian, strLen); err != nil {
		return nil, fmt.Errorf("failed to serialize Compression length: %v", err)
	}
	if _, err := buf.Write([]byte(str)); err != nil {
		return nil, fmt.Errorf("failed to serialize Compression: %v", err)
	}

	return buf.Bytes(), nil
}

// Deserialize populates the Metadata struct from a byte slice using binary decoding.
// It includes length validation for variable-length fields to prevent panics.
func (m *Metadata) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	fixedFields := []interface{}{
		&m.Version,
		&m.NumRecords,
		&m.DataSize,
		&m.IndexOffset,
		&m.IndexSize,
		&m.IndexChecksum,
		&m.BloomFilterOffset,
		&m.BloomFilterSize,
		&m.Timestamp,
	}

	for _, field := range fixedFields {
		if err := binary.Read(buf, binary.BigEndian, field); err != nil {
			return fmt.Errorf("failed to deserialize: %v", err)
		}
	}

	var strLen int32
	if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
		return fmt.Errorf("failed to deserialize Compression length: %v", err)
	}
	if strLen < 0 || strLen > 1<<20 { // 1MB limit to prevent excessive allocation
		return fmt.Errorf("invalid Compression length: %d", strLen)
	}

	strBytes := make([]byte, strLen)
	if _, err := io.ReadFull(buf, strBytes); err != nil {
		return fmt.Errorf("failed to deserialize Compression: %v", err)
	}
	m.Compression = string(strBytes)

	return nil
}
