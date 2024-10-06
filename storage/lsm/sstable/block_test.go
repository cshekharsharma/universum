package sstable

import (
	"bufio"
	"bytes"
	"hash/crc32"
	"testing"
	"universum/dslib"
	"universum/entity"
	"universum/resp3"
	"universum/utils"
)

func TestNewBlock(t *testing.T) {
	block := NewBlock(64 * 1024)
	if block == nil {
		t.Fatal("expected a valid Block instance")
	}
	if block.maxSize != 64*1024 {
		t.Fatalf("expected maxSize to be 64KB, got %d", block.maxSize)
	}
}

func TestBlock_AddRecord(t *testing.T) {
	block := NewBlock(64 * 1024)
	bloom := dslib.NewBloomFilter(1000, 5)

	record := &entity.ScalarRecord{
		Value:  "testValue",
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: utils.GetCurrentEPochTime() + 1000,
	}

	err := block.AddRecord("testKey", record.ToMap(), bloom)
	if err != nil {
		t.Fatalf("failed to add record: %v", err)
	}

	if block.startKey != "testKey" {
		t.Fatalf("expected startKey to be testKey, got %s", block.startKey)
	}

	if !bloom.Exists("testKey") {
		t.Fatal("expected bloom filter to contain the added key")
	}
}

func TestBlock_SerializeBlock(t *testing.T) {
	block := NewBlock(64 * 1024)
	bloom := dslib.NewBloomFilter(1000, 5)

	record := &entity.ScalarRecord{
		Value:  "testValue",
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: utils.GetCurrentEPochTime() + 1000,
	}

	block.AddRecord("testKey", record.ToMap(), bloom)
	serialized, err := block.SerializeBlock()
	if err != nil {
		t.Fatalf("failed to serialize block: %v", err)
	}

	if len(serialized) == 0 {
		t.Fatal("expected serialized block to have data")
	}

	if !block.ValidateBlock() {
		t.Fatal("block checksum validation failed")
	}
}

func TestBlock_ReadRecordAtOffset(t *testing.T) {
	block := NewBlock(64 * 1024)
	bloom := dslib.NewBloomFilter(1000, 5)

	record := &entity.ScalarRecord{
		Value:  "testValue",
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: utils.GetCurrentEPochTime() + 1000,
	}

	block.AddRecord("testKey", record.ToMap(), bloom)
	serialized, err := block.SerializeBlock()
	if err != nil {
		t.Fatalf("failed to serialize block: %v", err)
	}

	block.data = serialized
	key, value, err := block.ReadRecordAtOffset(0)
	if err != nil {
		t.Fatalf("failed to read record from block: %v", err)
	}

	if string(key) != "testKey" {
		t.Fatalf("expected key to be testKey, got %s", string(key))
	}

	decodedValue, err := resp3.Decode(bufio.NewReader(bytes.NewReader(value)))
	if err != nil {
		t.Fatalf("failed to decode value: %v", err)
	}

	if _, ok := decodedValue.(map[string]interface{}); ok {
		record, _ := decodedValue.(map[string]interface{})
		valFields := record["Value"]
		if valFields != "testValue" {
			t.Fatalf("expected value to be testValue, got %v", valFields)
		}
	} else {
		t.Fatalf("expected decoded value to be a map, got %T", decodedValue)
	}
}

func TestBlock_RemainingSpace(t *testing.T) {
	block := NewBlock(64 * 1024)
	remainingSpace := block.RemainingSpace()

	if remainingSpace != 64*1024 {
		t.Fatalf("expected remaining space to be 64KB, got %d", remainingSpace)
	}

	record := &entity.ScalarRecord{
		Value:  "testValue",
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: utils.GetCurrentEPochTime() + 1000,
	}

	bloom := dslib.NewBloomFilter(1000, 5)
	block.AddRecord("testKey", record.ToMap(), bloom)
	if block.RemainingSpace() >= remainingSpace {
		t.Fatalf("expected remaining space to decrease after adding a record")
	}
}

func TestBlock_ValidateBlock(t *testing.T) {
	block := NewBlock(64 * 1024)
	bloom := dslib.NewBloomFilter(1000, 5)

	record := &entity.ScalarRecord{
		Value:  "testValue",
		LAT:    utils.GetCurrentEPochTime(),
		Expiry: utils.GetCurrentEPochTime() + 1000,
	}

	block.AddRecord("testKey", record.ToMap(), bloom)
	serialized, err := block.SerializeBlock()
	if err != nil {
		t.Fatalf("failed to serialize block: %v", err)
	}

	serialized[0] = 0xFF
	block.data = serialized

	if block.ValidateBlock() {
		t.Fatal("expected block validation to fail with tampered data")
	}

	block.checksum = crc32.ChecksumIEEE(block.data)
	if !block.ValidateBlock() {
		t.Fatal("expected block validation to succeed after fixing checksum")
	}
}
