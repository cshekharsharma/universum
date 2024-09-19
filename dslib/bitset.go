package dslib

import (
	"errors"
)

// BitSet represents a set of bits, using an underlying slice of uint64.
type BitSet struct {
	Bits []uint64
	Size uint64
}

// NewBitSet creates a new BitSet with the specified number of bits.
func NewBitSet(size uint64) *BitSet {
	numWords := (size + 63) / 64
	return &BitSet{
		Bits: make([]uint64, numWords),
		Size: size,
	}
}

// Set sets the bit at the given position to 1.
func (bs *BitSet) Set(pos uint64) error {
	if pos >= bs.Size {
		return errors.New("position out of bounds")
	}
	word, bit := pos/64, pos%64
	bs.Bits[word] |= 1 << bit
	return nil
}

// Clear sets the bit at the given position to 0.
func (bs *BitSet) Clear(pos uint64) error {
	if pos >= bs.Size {
		return errors.New("position out of bounds")
	}
	word, bit := pos/64, pos%64
	bs.Bits[word] &^= 1 << bit
	return nil
}

// Efficiently clear all the blocks by resetting the underlying slice.
func (b *BitSet) ClearAll() {
	copy(b.Bits, make([]uint64, len(b.Bits)))
}

// IsSet checks whether the bit at the given position is set (1).
func (bs *BitSet) IsSet(pos uint64) (bool, error) {
	if pos >= bs.Size {
		return false, errors.New("position out of bounds")
	}
	word, bit := pos/64, pos%64
	return (bs.Bits[word]&(1<<bit) != 0), nil
}

// Toggle flips the bit at the given position.
func (bs *BitSet) Toggle(pos uint64) error {
	if pos >= bs.Size {
		return errors.New("position out of bounds")
	}
	word, bit := pos/64, pos%64
	bs.Bits[word] ^= 1 << bit // XOR to toggle
	return nil
}

// CountSetBits counts the number of bits that are set to 1.
func (bs *BitSet) CountSetBits() uint64 {
	var count uint64
	for _, word := range bs.Bits {
		count += uint64(popCount(word))
	}
	return count
}

// GetSize returns the total number of bits in the bitset.
func (bs *BitSet) GetSize() uint64 {
	return bs.Size
}

// popCount returns the number of set bits in a uint64.
func popCount(x uint64) int {
	// Kernighan's algorithm to count set bits
	count := 0
	for x != 0 {
		x &= (x - 1)
		count++
	}
	return count
}

// AllSet checks if all bits in the bitset are set to 1.
func (bs *BitSet) AllSet() bool {
	for i := 0; i < len(bs.Bits); i++ {
		if bs.Bits[i] != ^uint64(0) && uint64(i)*64 < bs.Size {
			return false
		}
	}
	return true
}

// AnySet checks if any bit in the bitset is set to 1.
func (bs *BitSet) AnySet() bool {
	for _, word := range bs.Bits {
		if word != 0 {
			return true
		}
	}
	return false
}

// NoneSet checks if no bits in the bitset are set (i.e., all bits are 0).
func (bs *BitSet) NoneSet() bool {
	for _, word := range bs.Bits {
		if word != 0 {
			return false
		}
	}
	return true
}

// String representation of the BitSet (for debugging purposes).
func (bs *BitSet) String() string {
	result := ""
	for i := uint64(0); i < bs.Size; i++ {
		if isSet, _ := bs.IsSet(i); isSet {
			result += "1"
		} else {
			result += "0"
		}
	}
	return result
}
