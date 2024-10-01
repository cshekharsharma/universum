package sstable

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestMetadataSerialization(t *testing.T) {
	original := Metadata{
		Version:           1,
		NumRecords:        1000,
		IndexOffset:       2048,
		DataSize:          102400,
		IndexSize:         5120,
		BloomFilterOffset: 1536,
		BloomFilterSize:   256,
		Timestamp:         1638307200,
		Compression:       "gzip",
		Checksum:          1234567890,
		IndexChecksum:     987654321,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	deserialized := Metadata{}
	err = deserialized.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Deserialized object does not match original.\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
	}
}

func TestMetadataSerializationEmptyFields(t *testing.T) {
	original := Metadata{}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	deserialized := Metadata{}
	err = deserialized.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Deserialized object does not match original.\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
	}
}

func TestMetadataSerializationPartialData(t *testing.T) {
	original := Metadata{
		Version:     1,
		NumRecords:  1000,
		Compression: "snappy",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	corruptedData := data[:len(data)-4]

	deserialized := Metadata{}
	err = deserialized.Deserialize(corruptedData)
	if err == nil {
		t.Fatalf("Expected deserialization to fail due to insufficient data")
	}
}

func TestMetadataSerializationInvalidData(t *testing.T) {
	invalidData := []byte("invalid data")

	deserialized := Metadata{}
	err := deserialized.Deserialize(invalidData)
	if err == nil {
		t.Fatalf("Expected deserialization to fail due to invalid data")
	}
}

func TestMetadataSerializationDifferentEndianness(t *testing.T) {
	original := Metadata{
		Version:           1,
		NumRecords:        1000,
		IndexOffset:       2048,
		DataSize:          102400,
		IndexSize:         5120,
		BloomFilterOffset: 1536,
		BloomFilterSize:   256,
		Timestamp:         1638307200,
		Compression:       "gzip",
		Checksum:          1234567890,
		IndexChecksum:     987654321,
	}

	buf := new(bytes.Buffer)
	fixedFields := []interface{}{
		original.Version,
		original.NumRecords,
		original.IndexOffset,
		original.DataSize,
		original.IndexSize,
		original.BloomFilterOffset,
		original.BloomFilterSize,
		original.Timestamp,
		original.Checksum,
		original.IndexChecksum,
	}

	for _, field := range fixedFields {
		if err := binary.Write(buf, binary.LittleEndian, field); err != nil {
			t.Fatalf("failed to serialize: %v", err)
		}
	}

	str := original.Compression
	strLen := int32(len(str))
	if err := binary.Write(buf, binary.LittleEndian, strLen); err != nil {
		t.Fatalf("failed to serialize Compression length: %v", err)
	}
	if _, err := buf.Write([]byte(str)); err != nil {
		t.Fatalf("failed to serialize Compression: %v", err)
	}

	data := buf.Bytes()

	deserialized := Metadata{}
	err := deserialized.Deserialize(data)
	if err == nil {
		t.Fatalf("Expected deserialization to fail due to endianness mismatch")
	}
}

func TestMetadataSerializationInvalidStringLength(t *testing.T) {
	original := Metadata{
		Version:     1,
		NumRecords:  1000,
		Compression: "gzip",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	data[len(data)-len(original.Compression)-4] = 0xFF // Set length to negative value

	deserialized := Metadata{}
	err = deserialized.Deserialize(data)
	if err == nil {
		t.Fatalf("Expected deserialization to fail due to invalid string length")
	}
}

func TestMetadataSerializationExcessiveStringLength(t *testing.T) {
	original := Metadata{
		Version:     1,
		NumRecords:  1000,
		Compression: "gzip",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	data[len(data)-len(original.Compression)-4] = 0x7F // Set length to a large positive value

	deserialized := Metadata{}
	err = deserialized.Deserialize(data)
	if err == nil {
		t.Fatalf("Expected deserialization to fail due to excessive string length")
	}
}
