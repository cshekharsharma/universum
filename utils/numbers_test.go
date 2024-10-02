package utils

import (
	"reflect"
	"testing"
)

func TestIsNumber(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{0, true},
		{int8(0), true},
		{int16(0), true},
		{int32(0), true},
		{int64(0), true},
		{uint8(0), true},
		{uint16(0), true},
		{uint32(0), true},
		{uint64(0), true},
		{float32(0), true},
		{float64(0), true},
		{"not a number", false},
		{true, false},
	}

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).Kind().String(), func(t *testing.T) {
			if got := IsNumber(tt.value); got != tt.want {
				t.Errorf("IsNumber() = %v, want %v for type %v", got, tt.want, reflect.TypeOf(tt.value).Kind())
			}
		})
	}
}

func TestIsInteger(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{0, true},
		{int8(0), true},
		{int16(0), true},
		{int32(0), true},
		{int64(0), true},
		{uint8(0), true},
		{uint16(0), true},
		{uint32(0), true},
		{uint64(0), true},
		{float32(0), false},
		{float64(0), false},
		{"not a number", false},
		{true, false},
	}

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).Kind().String(), func(t *testing.T) {
			if got := IsInteger(tt.value); got != tt.want {
				t.Errorf("IsInteger() = %v, want %v for type %v", got, tt.want, reflect.TypeOf(tt.value).Kind())
			}
		})
	}
}

func TestIsFloat(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{0, false},
		{int8(0), false},
		{int16(0), false},
		{int32(0), false},
		{int64(0), false},
		{uint8(0), false},
		{uint16(0), false},
		{uint32(0), false},
		{uint64(0), false},
		{float32(0), true},
		{float64(0), true},
		{"not a number", false},
		{true, false},
	}

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).Kind().String(), func(t *testing.T) {
			if got := IsFloat(tt.value); got != tt.want {
				t.Errorf("IsFloat() = %v, want %v for type %v", got, tt.want, reflect.TypeOf(tt.value).Kind())
			}
		})
	}
}

func TestMaxUint64(t *testing.T) {
	tests := []struct {
		a, b uint
		want uint
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, 1},
		{1, 1, 1},
		{1, 2, 2},
		{2, 1, 2},
		{2, 2, 2},
		{2, 3, 3},
		{3, 2, 3},
		{3, 3, 3},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := MaxUint64(uint64(tt.a), uint64(tt.b)); got != uint64(tt.want) {
				t.Errorf("MaxUint64() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := MaxUint8(uint8(tt.a), uint8(tt.b)); got != uint8(tt.want) {
				t.Errorf("MaxUint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackAndUnpackNumbers(t *testing.T) {
	tests := []struct {
		num1 int32
		num2 int32
	}{
		{123456789, 987654321},
		{-123456789, 987654321},
		{123456789, -987654321},
		{-123456789, -987654321},
		{0, 0},
		{1, -1},
	}

	for _, test := range tests {
		packed := PackNumbers(test.num1, test.num2)

		unpackedNum1, unpackedNum2 := UnpackNumbers(packed)
		if unpackedNum1 != test.num1 || unpackedNum2 != test.num2 {
			t.Errorf("UnpackNumbers(%d, %d) = %d, %d; want %d, %d", test.num1, test.num2, unpackedNum1, unpackedNum2, test.num1, test.num2)
		}

		firstNum := UnpackFirstNumber(packed)
		if firstNum != test.num1 {
			t.Errorf("UnpackFirstNumber(%d, %d) = %d; want %d", test.num1, test.num2, firstNum, test.num1)
		}

		secondNum := UnpackSecondNumber(packed)
		if secondNum != test.num2 {
			t.Errorf("UnpackSecondNumber(%d, %d) = %d; want %d", test.num1, test.num2, secondNum, test.num2)
		}
	}
}
