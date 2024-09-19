package dslib

import (
	"testing"
)

func TestNewBitSet(t *testing.T) {
	bs := NewBitSet(128)

	if bs.Size != 128 {
		t.Errorf("expected size to be 128, got %d", bs.Size)
	}

	if len(bs.Bits) != 2 { // 128 bits means 2 uint64 words (64 bits per word)
		t.Errorf("expected 2 uint64 words, got %d", len(bs.Bits))
	}
}

func TestSetAndIsSet(t *testing.T) {
	bs := NewBitSet(64)

	if err := bs.Set(10); err != nil {
		t.Errorf("unexpected error while setting bit: %v", err)
	}

	isSet, err := bs.IsSet(10)
	if err != nil {
		t.Errorf("unexpected error in IsSet: %v", err)
	}

	if !isSet {
		t.Errorf("expected bit 10 to be set")
	}

	isSet, err = bs.IsSet(11)
	if err != nil {
		t.Errorf("unexpected error in IsSet: %v", err)
	}

	if isSet {
		t.Errorf("expected bit 11 to be not set")
	}
}

func TestSetOutOfBounds(t *testing.T) {
	bs := NewBitSet(64)

	err := bs.Set(65)
	if err == nil {
		t.Errorf("expected error for out of bounds set, got none")
	}
}

func TestClearAndIsSet(t *testing.T) {
	bs := NewBitSet(64)
	bs.Set(10)

	if err := bs.Clear(10); err != nil {
		t.Errorf("unexpected error while clearing bit: %v", err)
	}

	isSet, err := bs.IsSet(10)
	if err != nil {
		t.Errorf("unexpected error in IsSet: %v", err)
	}

	if isSet {
		t.Errorf("expected bit 10 to be cleared")
	}
}

func TestToggle(t *testing.T) {
	bs := NewBitSet(64)

	if err := bs.Toggle(20); err != nil {
		t.Errorf("unexpected error in Toggle: %v", err)
	}

	isSet, err := bs.IsSet(20)
	if err != nil {
		t.Errorf("unexpected error in IsSet: %v", err)
	}

	if !isSet {
		t.Errorf("expected bit 20 to be set after toggle")
	}

	if err := bs.Toggle(20); err != nil {
		t.Errorf("unexpected error in Toggle: %v", err)
	}

	isSet, err = bs.IsSet(20)
	if err != nil {
		t.Errorf("unexpected error in IsSet: %v", err)
	}

	if isSet {
		t.Errorf("expected bit 20 to be cleared after second toggle")
	}
}

func TestCountSetBits(t *testing.T) {
	bs := NewBitSet(64)
	bs.Set(10)
	bs.Set(20)
	bs.Set(30)

	if count := bs.CountSetBits(); count != 3 {
		t.Errorf("expected 3 set bits, got %d", count)
	}
}

func TestClearAll(t *testing.T) {
	bs := NewBitSet(64)
	bs.Set(10)
	bs.Set(20)
	bs.ClearAll()

	if bs.CountSetBits() != 0 {
		t.Errorf("expected 0 set bits after ClearAll")
	}
}

func TestAllSet(t *testing.T) {
	bs := NewBitSet(128)
	for i := uint64(0); i < 128; i++ {
		bs.Set(i)
	}

	if !bs.AllSet() {
		t.Errorf("expected all bits to be set")
	}
}

func TestAnySet(t *testing.T) {
	bs := NewBitSet(64)

	if bs.AnySet() {
		t.Errorf("expected no bits to be set initially")
	}

	bs.Set(0)
	if !bs.AnySet() {
		t.Errorf("expected at least one bit to be set")
	}
}

func TestNoneSet(t *testing.T) {
	bs := NewBitSet(64)

	if !bs.NoneSet() {
		t.Errorf("expected none of the bits to be set initially")
	}

	bs.Set(0)
	if bs.NoneSet() {
		t.Errorf("expected some bits to be set")
	}
}

func TestBitSetString(t *testing.T) {
	bs := NewBitSet(10)

	expected := "0000000000"
	if result := bs.String(); result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	bs.Set(1)
	bs.Set(3)
	bs.Set(5)

	expected = "0101010000"
	if result := bs.String(); result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	bs.Clear(3)
	expected = "0100010000"
	if result := bs.String(); result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
