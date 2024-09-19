package crypto

import (
	"encoding/binary"
)

const (
	c1 uint64 = 0x87c37b91114253d5
	c2 uint64 = 0x4cf5ad432745937f
)

// This implementation is adapted from publicly available sources of MurmurHash3.
// MurmurHash3 was created by Austin Appleby and is available under the MIT License.
// See: https://github.com/aappleby/smhasher
//
// MurmurHash is a non-cryptographic hash function designed for general hash-based lookups.
func MurmurHash64(data []byte, seed uint64) uint64 {
	h1 := seed
	length := len(data)
	nblocks := length / 16

	// Process the data in 16-byte blocks
	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint64(data[i*16 : i*16+8])
		k2 := binary.LittleEndian.Uint64(data[i*16+8 : i*16+16])

		k1 *= c1
		k1 = rotl64(k1, 31)
		k1 *= c2
		h1 ^= k1

		h1 = rotl64(h1, 27)
		h1 = h1*5 + 0x52dce729

		k2 *= c2
		k2 = rotl64(k2, 33)
		k2 *= c1
		h1 ^= k2
	}

	// Process the tail (remaining data < 16 bytes)
	var k1, k2 uint64
	tail := data[nblocks*16:]

	switch len(tail) & 15 {
	case 15:
		k2 ^= uint64(tail[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tail[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tail[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tail[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tail[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tail[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tail[8])
		k2 *= c2
		k2 = rotl64(k2, 33)
		k2 *= c1
		h1 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(tail[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0])
		k1 *= c1
		k1 = rotl64(k1, 31)
		k1 *= c2
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint64(length)
	h1 = fmix64(h1)

	return h1
}

func rotl64(x, r uint64) uint64 {
	return (x << r) | (x >> (64 - r))
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}
