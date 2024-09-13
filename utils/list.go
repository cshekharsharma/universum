package utils

import "reflect"

// ExistsInList checks if a given value exists within a slice and returns
// a boolean indicating its existence and the index at which it was found.
//
// Parameters:
//   - val: The value to search for in the slice.
//   - array: The slice to search within.
//
// Returns:
//   - exists: A boolean value indicating whether 'val' exists in 'array'.
//   - index: An integer representing the index where 'val' was found in 'array'.
//     If 'val' is not found, the index is set to -1.
func ExistsInList(val interface{}, array interface{}) (bool, int) {
	exists := false
	index := -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				return true, index
			}
		}
	}

	return exists, index
}
