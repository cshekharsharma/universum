package dslib

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"universum/crypto"
	"universum/utils"
)

var primeSeeds = []uint64{
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
	73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151,
	157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229, 233,
	239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317,
	331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409, 419,
	421, 431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499, 503,
	509, 521, 523, 541, 547, 557, 563, 569, 571, 577, 587, 593, 599, 601, 607,
}

type BloomFilter struct {
	Size      uint64
	HashCount uint8
	Bitset    *BitSet
}

func NewBloomFilter(size uint64, hashCount uint8) *BloomFilter {
	return &BloomFilter{
		Size:      utils.MaxUint64(size, 1),
		HashCount: utils.MaxUint8(hashCount, 1),
		Bitset:    NewBitSet(size),
	}
}

func NewBloomFilterWithPositiveRate(size uint64, rate float64) *BloomFilter {
	if rate <= 0 || rate >= 1 {
		rate = 0.01
	}

	n := float64(size) / (-math.Log(rate) / math.Pow(math.Ln2, 2))
	hashCount := uint8(math.Ceil((float64(size) / n) * math.Ln2))

	return &BloomFilter{
		Size:      utils.MaxUint64(size, 1),
		HashCount: utils.MaxUint8(hashCount, 1),
		Bitset:    NewBitSet(size),
	}
}

func (bf *BloomFilter) Add(key string) {
	for i := uint8(0); i < bf.HashCount; i++ {
		seed := primeSeeds[i%bf.HashCount]

		position := bf.Hash(key, seed)
		bf.Bitset.Set(position)
	}
}

func (bf *BloomFilter) Exists(key string) bool {
	for i := uint8(0); i < bf.HashCount; i++ {
		seed := primeSeeds[i%bf.HashCount]

		position := bf.Hash(key, seed)
		exists, err := bf.Bitset.IsSet(position)

		if err != nil {
			return false
		}

		if !exists {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) ClearAll() {
	bf.Bitset.ClearAll()
}

func (bf *BloomFilter) Hash(key string, seed uint64) uint64 {
	hash := crypto.MurmurHash64([]byte(key), seed)
	return hash % bf.Size
}

func (bf *BloomFilter) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(bf)

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes the Bloom filter from a byte slice.
func (bf *BloomFilter) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(bf)

	if err != nil {
		return err
	}
	return nil
}

func (bf *BloomFilter) Merge(other *BloomFilter) error {
	if bf.Size != other.Size || bf.HashCount != other.HashCount {
		return fmt.Errorf("bloom filters must have the same size and hash count to be merged")
	}

	// Merge bitsets by performing a bitwise OR on both
	for i := uint64(0); i < bf.Size; i++ {
		exists, err := other.Bitset.IsSet(i)
		if err != nil {
			exists = false
		}
		if exists {
			bf.Bitset.Set(i)
		}
	}
	return nil
}

func (bf *BloomFilter) MemoryUsage() uint64 {
	return bf.Size / 8
}

func OptimalBloomFilterSize(numKeys uint64, falsePositiveRate float64) (uint64, uint8) {
	n := float64(numKeys)
	m := (-n * math.Log(falsePositiveRate)) / math.Pow(math.Ln2, 2)
	k := uint8(math.Ceil(math.Ln2 * (m / n)))

	return uint64(math.Ceil(m)), k
}
