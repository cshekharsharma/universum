package utils

import (
	"testing"
)

func TestExistsInList(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		val           interface{}
		array         interface{}
		expected      bool
		expectedIndex int
	}{
		{
			name:          "ElementExistsInSlice",
			val:           3,
			array:         []int{1, 2, 3, 4, 5},
			expected:      true,
			expectedIndex: 2,
		},
		{
			name:          "ElementNotExistInSlice",
			val:           "test",
			array:         []string{"one", "two", "three"},
			expected:      false,
			expectedIndex: -1,
		},
		{
			name:          "EmptySlice",
			val:           42,
			array:         []int{},
			expected:      false,
			expectedIndex: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, index := ExistsInList(tc.val, tc.array)

			if exists != tc.expected {
				t.Errorf("Test %s failed: expected exists to be %v, got %v", tc.name, tc.expected, exists)
			}
			if index != tc.expectedIndex {
				t.Errorf("Test %s failed: expected index to be %v, got %v", tc.name, tc.expectedIndex, index)
			}
		})
	}
}
