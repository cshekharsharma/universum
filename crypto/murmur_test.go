package crypto

import (
	"testing"
)

func TestMurmurHash64(t *testing.T) {
	tests := []struct {
		data []byte
		seed uint64
	}{
		{
			data: []byte("hello"),
			seed: 12345678,
		},
		{
			data: []byte("world"),
			seed: 87654321,
		},
		{
			data: []byte("MurmurHash64"),
			seed: 987654321,
		},
		{
			data: []byte(""),
			seed: 0,
		},
	}

	for _, tt := range tests {
		result1 := MurmurHash64(tt.data, tt.seed)
		result2 := MurmurHash64(tt.data, tt.seed)
		if result1 != result2 {
			t.Errorf("MurmurHash64(%q, %d) = %x, but MurmurHash64 returned different result %x", tt.data, tt.seed, result1, result2)
		}
	}
}

func TestMurmurHash64WithDifferentSeeds(t *testing.T) {
	data := []byte("test")
	seeds := []uint64{0, 1, 123, 456, 789, 987654321}

	for _, seed := range seeds {
		result := MurmurHash64(data, seed)
		if result == 0 {
			t.Errorf("MurmurHash64(%q, %d) resulted in zero hash", data, seed)
		}
	}
}

func TestMurmurHash64EmptyString(t *testing.T) {
	result := MurmurHash64([]byte(""), 12345)
	if result == 0 {
		t.Error("MurmurHash64 with empty string resulted in zero hash")
	}
}
