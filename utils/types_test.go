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
