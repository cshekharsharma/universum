package utils

import "reflect"

var numberTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
}

var integerTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
}

var floatTypes = []reflect.Kind{
	reflect.Float32,
	reflect.Float64,
}

// Checks if the provided variable is a number.
// The number can be any of type, ie signed int, unsigned int or float.
func IsNumber(value interface{}) bool {
	datatype := reflect.TypeOf(value).Kind()

	var isNum bool = false
	for i := range numberTypes {
		if numberTypes[i] == datatype {
			isNum = true
			break
		}
	}

	return isNum
}

// Checks if the provided variable is an integer.
// The integer can be any of type, ie signed int or unsigned int of any bitsize.
func IsInteger(value interface{}) bool {
	datatype := reflect.TypeOf(value).Kind()

	var isInt bool = false
	for i := range integerTypes {
		if integerTypes[i] == datatype {
			isInt = true
			break
		}
	}

	return isInt
}

// Checks if the provided variable is an float.
// The float can be any of any bitsize.
func IsFloat(value interface{}) bool {
	datatype := reflect.TypeOf(value).Kind()

	var isFloat bool = false
	for i := range floatTypes {
		if floatTypes[i] == datatype {
			isFloat = true
			break
		}
	}

	return isFloat
}
