package utils

import "testing"

func TestGetTypeEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint8
	}{
		{"NilType", nil, TYPE_ENCODING_NIL},
		{"BoolTrue", true, TYPE_ENCODING_BOOL},
		{"BoolFalse", false, TYPE_ENCODING_BOOL},
		{"Integer", 123, TYPE_ENCODING_INT},
		{"NegativeInteger", -456, TYPE_ENCODING_INT},
		{"Float", 3.14, TYPE_ENCODING_FLOAT},
		{"NegativeFloat", -1.23, TYPE_ENCODING_FLOAT},
		{"String", "hello", TYPE_ENCODING_STRING},
		{"EmptyString", "", TYPE_ENCODING_STRING},
		{"Array", [3]int{1, 2, 3}, TYPE_ENCODING_ARRAY},
		{"Slice", []int{1, 2, 3}, TYPE_ENCODING_SLICE},
		{"Map", map[string]int{"a": 1, "b": 2}, TYPE_ENCODING_MAP},
		{"Raw", struct{}{}, TYPE_ENCODING_RAW},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTypeEncoding(tt.input)
			if got != tt.expected {
				t.Errorf("GetTypeEncoding(%v) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsWriteableDatatype(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{name: "StringType", value: "hello", expected: true},
		{name: "IntType", value: 123, expected: true},
		{name: "FloatType", value: 12.34, expected: true},
		{name: "BoolType", value: true, expected: true},
		{name: "UnsupportedType", value: struct{}{}, expected: false},
		{name: "PointerType", value: &struct{}{}, expected: false},
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
		{name: "StringBelowMaxSize", value: "short", expected: true},
		{name: "StringEqualsMaxSize", value: "abcdefghij", expected: true},
		{name: "StringExceedsMaxSize", value: "exceeding size limit", expected: false},
		{name: "NonStringTypeInt", value: 123, expected: true},
		{name: "NonStringTypeFloat", value: 123.45, expected: true},
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
