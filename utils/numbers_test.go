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
