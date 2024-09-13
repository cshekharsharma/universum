package utils

import (
	"testing"
)

func TestIsWriteableDatatype(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{name: "String Type", value: "hello", expected: true},
		{name: "Int Type", value: 123, expected: true},
		{name: "Float Type", value: 12.34, expected: true},
		{name: "Bool Type", value: true, expected: true},
		{name: "Unsupported Type", value: struct{}{}, expected: false},
		{name: "Pointer Type", value: &struct{}{}, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsWriteableDatatype(test.value)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestIsWriteableDataSize(t *testing.T) {
	var maxsize int64 = 10

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{name: "String below max size", value: "short", expected: true},
		{name: "String exactly max size", value: "abcdefghij", expected: true},
		{name: "String exceeds max size", value: "exceeding size limit", expected: false},
		{name: "Non-string type (int)", value: 123, expected: true},
		{name: "Non-string type (float)", value: 123.45, expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsWriteableDataSize(test.value, maxsize)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
