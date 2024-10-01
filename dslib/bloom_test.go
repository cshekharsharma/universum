package dslib

import (
	"strconv"
	"testing"
)

func TestNewBloomFilterWithPositiveRate(t *testing.T) {
	tests := []struct {
		name            string
		size            uint64
		rate            float64
		expectedRate    float64
		expectedHashCnt uint8
	}{
		{
			name:            "ValidRate",
			size:            1000,
			rate:            0.05,
			expectedRate:    0.05,
			expectedHashCnt: 5,
		},
		{
			name:            "RateTooLow",
			size:            1000,
			rate:            -0.01,
			expectedRate:    0.01,
			expectedHashCnt: 7,
		},
		{
			name:            "RateTooHigh",
			size:            1000,
			rate:            1.2,
			expectedRate:    0.01,
			expectedHashCnt: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewBloomFilterWithPositiveRate(tt.size, tt.rate)

			if bf.Size != tt.size {
				t.Errorf("expected size %v, got %v", tt.size, bf.Size)
			}

			if bf.HashCount != tt.expectedHashCnt {
				t.Errorf("expected hash count %v, got %v", tt.expectedHashCnt, bf.HashCount)
			}

			if bf.Bitset.Size != tt.size {
				t.Errorf("expected bitset size %v, got %v", tt.size, bf.Bitset.Size)
			}
		})
	}
}

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(1000, 5)

	bf.Add("hello")
	if !bf.Exists("hello") {
		t.Errorf("Expected 'hello' to exist in the Bloom filter")
	}

	if bf.Exists("world") {
		t.Errorf("Expected 'world' to not exist in the Bloom filter")
	}

	bf.ClearAll()
	if bf.Exists("hello") {
		t.Errorf("Expected 'hello' to be removed after ClearAll")
	}
}

func TestBloomFilterSerialization(t *testing.T) {
	bf := NewBloomFilter(1000, 5)

	bf.Add("test")
	data, err := bf.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize Bloom filter: %v", err)
	}

	newBF := NewBloomFilter(1000, 5)
	err = newBF.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize Bloom filter: %v", err)
	}

	if !newBF.Exists("test") {
		t.Errorf("Expected 'test' to exist after deserialization")
	}
}

func TestBloomFilterMerge(t *testing.T) {
	bf1 := NewBloomFilter(1000, 5)
	bf2 := NewBloomFilter(1000, 5)

	bf1.Add("foo")
	bf2.Add("bar")

	err := bf1.Merge(bf2)
	if err != nil {
		t.Fatalf("Failed to merge Bloom filters: %v", err)
	}

	if !bf1.Exists("foo") {
		t.Errorf("Expected 'foo' to exist after merge")
	}
	if !bf1.Exists("bar") {
		t.Errorf("Expected 'bar' to exist after merge")
	}
}

func TestOptimalBloomFilterSize(t *testing.T) {
	numKeys := int64(1000)
	falsePositiveRate := 0.01

	size, hashCount := OptimalBloomFilterSize(numKeys, falsePositiveRate)

	if size == 0 || hashCount == 0 {
		t.Errorf("Expected valid size and hash count, got size=%d, hashCount=%d", size, hashCount)
	}

	bf := NewBloomFilter(size, hashCount)
	for i := 0; i < int(numKeys); i++ {
		bf.Add(strconv.Itoa(i))
	}
	for i := 0; i < int(numKeys); i++ {
		if !bf.Exists(strconv.Itoa(i)) {
			t.Errorf("Expected key %d to exist in the Bloom filter", i)
		}
	}
}

func TestMemoryUsage(t *testing.T) {
	bf := NewBloomFilter(1000, 5)
	usage := bf.MemoryUsage()
	expectedUsage := uint64(1000 / 8) // Each bit takes 1 bit, so 1000 bits = 125 bytes

	if usage != expectedUsage {
		t.Errorf("Expected memory usage to be %d, got %d", expectedUsage, usage)
	}
}

func TestSerializeDeserialize(t *testing.T) {
	bf := NewBloomFilter(1000, 5)
	bf.Add("example")

	data, err := bf.Serialize()
	if err != nil {
		t.Fatalf("Error during serialization: %v", err)
	}

	newBF := NewBloomFilter(1000, 5)
	err = newBF.Deserialize(data)
	if err != nil {
		t.Fatalf("Error during deserialization: %v", err)
	}

	if !newBF.Exists("example") {
		t.Errorf("Expected 'example' to exist after deserialization")
	}
}
