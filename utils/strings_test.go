package utils

import (
	"fmt"
	"testing"
	"time"
	"unicode"
)

func TestIsString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"String input", "hello", true},
		{"Integer input", 123, false},
		{"Float input", 123.45, false},
		{"Boolean input", true, false},
		{"Byte slice", []byte("byte slice"), false},
		{"Struct input", struct{}{}, false},
		{"Nil input", nil, false},
		{"Function input", func() {}, false},
		{"Rune input", 'a', false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsString(test.input)
			if result != test.expected {
				t.Errorf("IsString(%v) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}

func TestGetRandomString(t *testing.T) {
	lengths := []int64{0, 1, 5, 10, 50, 100}

	for _, length := range lengths {
		t.Run("Length_"+fmt.Sprintf("%d", length), func(t *testing.T) {
			str := GetRandomString(length)
			if int64(len(str)) != length {
				t.Errorf("GetRandomString(%d) returned string of length %d; want %d", length, len(str), length)
			}

			for _, r := range str {
				if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
					t.Errorf("GetRandomString(%d) returned invalid character '%c'", length, r)
				}
			}
		})
	}

	str1 := GetRandomString(10)
	time.Sleep(2 * time.Millisecond)
	str2 := GetRandomString(10)
	if str1 == str2 {
		t.Errorf("GetRandomString(10) produced the same string '%s' twice; expected different strings", str1)
	}
}

func TestGetRandomStringCrypto_ValidLength(t *testing.T) {
	length := uint8(10)
	result, err := GetRandomStringCrypto(length)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != int(length) {
		t.Errorf("Expected length %d, got %d", length, len(result))
	}
}

func TestGetRandomStringCrypto_ZeroLength(t *testing.T) {
	length := uint8(0)
	result, err := GetRandomStringCrypto(length)
	if err != nil {
		t.Errorf("Expected no error for length 0, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty string, got %d characters", len(result))
	}
}
