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

// MaxUint64 returns the maximum of two uint32 numbers.
func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MaxUint8 returns the maximum of two uint8 numbers.
func MaxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func PackNumbers(num1 int32, num2 int32) int64 {
	return (int64(int32(num1)) << 32) | int64(uint32(num2))
}

// Unpack both numbers from the packed uint64
func UnpackNumbers(packed int64) (int32, int32) {
	num1 := int32(packed >> 32)
	num2 := int32(packed & 0xFFFFFFFF)
	return num1, num2
}

// Get only the first number from the packed uint64
func UnpackFirstNumber(packed int64) int32 {
	return int32(packed >> 32)
}

// Get only the second number from the packed uint64
func UnpackSecondNumber(packed int64) int32 {
	return int32(packed & 0xFFFFFFFF)
}
