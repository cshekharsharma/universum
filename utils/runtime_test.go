package utils

import (
	"testing"
)

func TestGetSizeInBytes(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedSize int64
	}{
		{"TestInt", 42, 8},
		{"TestInt8", int8(42), 1},
		{"TestInt16", int16(42), 2},
		{"TestInt32", int32(42), 4},
		{"TestInt64", int64(42), 8},
		{"TestUint", uint(42), 8},
		{"TestUint8", uint8(42), 1},
		{"TestUint16", uint16(42), 2},
		{"TestUint32", uint32(42), 4},
		{"TestUint64", uint64(42), 8},
		{"TestFloat32", float32(3.14), 4},
		{"TestFloat64", float64(3.14), 8},
		{"TestBool", true, 1},

		{"TestString", "hello", int64(16 + len("hello"))},
		{"TestEmpty", "", 16},

		{"TestIntSlice", []int{1, 2, 3, 4, 5}, 24 + 5*8},
		{"TestInt8Slice", []int8{1, 2, 3}, 24 + 3*1},
		{"TestBoolSlice", []bool{true, false, true}, 24 + 3},

		{"TestStringSlice", []string{"a", "bb", "hello world!"}, int64(24 + 16*3 + len("a") + len("bb") + len("hello world!"))},
		{"TestEmptyStringSlice", []string{}, 24},

		{"TestIntSliceSlice", [][]int{{1, 2}, {3, 4}}, 24 + 24 + 2*8 + 24 + 2*8},
		{"TestEmptyIntSlice", []int{}, 24},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualSize, err := GetSizeInBytes(tc.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if actualSize != tc.expectedSize {
				t.Errorf("Expected %d bytes, but got %d bytes", tc.expectedSize, actualSize)
			}
		})
	}
}
