package utils

import (
	"reflect"
	"testing"
)

func TestSortMapKeys(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "SortedKeysMap",
			input: map[string]interface{}{
				"b": 2,
				"a": 1,
				"c": 3,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			name: "AlreadySortedMap",
			input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			name:     "EmptyMap",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "SingleElementMap",
			input: map[string]interface{}{
				"z": 10,
			},
			expected: map[string]interface{}{
				"z": 10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := SortMapKeys(tc.input)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", tc.name, tc.expected, actual)
			}
		})
	}
}
