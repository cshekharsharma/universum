package config

import (
	"reflect"
	"testing"
)

type Owner struct {
	Name    string `toml:"name"`
	Age     int    `toml:"age"`
	Address string `toml:"address"`
}

type NestedOwner struct {
	Owner *Owner `toml:"owner"` // Pointer to struct
}

type TestConfig struct {
	Title   string  `toml:"title"`
	Enabled bool    `toml:"enabled"`
	Count   int     `toml:"count"`
	Pi      float64 `toml:"pi"`
	Owner   Owner   `toml:"owner"`
	Numbers []int   `toml:"numbers"`
}

// TestUnmarshalBasicTypes tests unmarshalling of basic TOML types into a struct
func TestUnmarshalBasicTypes(t *testing.T) {
	tomlData := TOMLTable{
		"title":   "Example TOML",
		"enabled": true,
		"count":   42,
		"pi":      3.14159,
	}

	var config TestConfig
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	if config.Title != "Example TOML" {
		t.Errorf("Expected title to be 'Example TOML', got %v", config.Title)
	}

	if config.Enabled != true {
		t.Errorf("Expected enabled to be true, got %v", config.Enabled)
	}

	if config.Count != 42 {
		t.Errorf("Expected count to be 42, got %v", config.Count)
	}

	if config.Pi != 3.14159 {
		t.Errorf("Expected pi to be 3.14159, got %v", config.Pi)
	}
}

// TestUnmarshalNestedStruct tests unmarshalling of nested TOML tables into Go structs
func TestUnmarshalNestedStruct(t *testing.T) {
	tomlData := TOMLTable{
		"owner": TOMLTable{
			"name":    "Tom",
			"age":     36,
			"address": "123 Main St",
		},
	}

	var config TestConfig
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	if config.Owner.Name != "Tom" {
		t.Errorf("Expected owner name to be 'Tom', got %v", config.Owner.Name)
	}

	if config.Owner.Age != 36 {
		t.Errorf("Expected owner age to be 36, got %v", config.Owner.Age)
	}

	if config.Owner.Address != "123 Main St" {
		t.Errorf("Expected owner address to be '123 Main St', got %v", config.Owner.Address)
	}
}

// TestUnmarshalPointerToStruct tests unmarshalling of pointers to structs
func TestUnmarshalPointerToStruct(t *testing.T) {
	tomlData := TOMLTable{
		"owner": TOMLTable{
			"name":    "Alice",
			"age":     29,
			"address": "456 Elm St",
		},
	}

	var config NestedOwner
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	if config.Owner == nil {
		t.Fatalf("Expected owner to be non-nil")
	}

	if config.Owner.Name != "Alice" {
		t.Errorf("Expected owner name to be 'Alice', got %v", config.Owner.Name)
	}

	if config.Owner.Age != 29 {
		t.Errorf("Expected owner age to be 29, got %v", config.Owner.Age)
	}

	if config.Owner.Address != "456 Elm St" {
		t.Errorf("Expected owner address to be '456 Elm St', got %v", config.Owner.Address)
	}
}

func TestUnmarshalArray(t *testing.T) {
	tomlData := TOMLTable{
		"numbers": []TOMLValue{1, 2, 3, 4, 5},
	}

	var config TestConfig
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	expectedNumbers := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(config.Numbers, expectedNumbers) {
		t.Errorf("Expected numbers to be %v, got %v", expectedNumbers, config.Numbers)
	}
}

func TestUnmarshalFullStruct(t *testing.T) {
	tomlData := TOMLTable{
		"title":   "Example TOML",
		"enabled": true,
		"count":   42,
		"pi":      3.14159,
		"owner": TOMLTable{
			"name":    "Tom",
			"age":     36,
			"address": "123 Main St",
		},
		"numbers": []TOMLValue{1, 2, 3, 4, 5},
	}

	var config TestConfig
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	if config.Title != "Example TOML" {
		t.Errorf("Expected title to be 'Example TOML', got %v", config.Title)
	}

	if config.Enabled != true {
		t.Errorf("Expected enabled to be true, got %v", config.Enabled)
	}

	if config.Count != 42 {
		t.Errorf("Expected count to be 42, got %v", config.Count)
	}

	if config.Pi != 3.14159 {
		t.Errorf("Expected pi to be 3.14159, got %v", config.Pi)
	}

	if config.Owner.Name != "Tom" {
		t.Errorf("Expected owner name to be 'Tom', got %v", config.Owner.Name)
	}

	if config.Owner.Age != 36 {
		t.Errorf("Expected owner age to be 36, got %v", config.Owner.Age)
	}

	if config.Owner.Address != "123 Main St" {
		t.Errorf("Expected owner address to be '123 Main St', got %v", config.Owner.Address)
	}

	expectedNumbers := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(config.Numbers, expectedNumbers) {
		t.Errorf("Expected numbers to be %v, got %v", expectedNumbers, config.Numbers)
	}
}

func TestUnmarshalMissingFields(t *testing.T) {
	tomlData := TOMLTable{
		"title": "Example TOML",
	}

	var config TestConfig
	err := Unmarshal(tomlData, &config)
	if err != nil {
		t.Fatalf("Error unmarshalling TOML: %v", err)
	}

	if config.Title != "Example TOML" {
		t.Errorf("Expected title to be 'Example TOML', got %v", config.Title)
	}

	if config.Enabled != false {
		t.Errorf("Expected enabled to be false, got %v", config.Enabled)
	}

	if config.Count != 0 {
		t.Errorf("Expected count to be 0, got %v", config.Count)
	}
}

// TestUnmarshalInvalidData tests handling of invalid data
func TestUnmarshalInvalidData(t *testing.T) {
	invalidData := TOMLTable{
		"count": "not_an_int", // This should cause an error
	}

	var config TestConfig
	err := Unmarshal(invalidData, &config)
	if err == nil {
		t.Fatalf("Expected error when unmarshalling invalid data, but got none")
	}

	expectedErr := "failed to set field count: expected int, got not_an_int"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %v, got: %v", expectedErr, err)
	}
}

func TestUnmarshalUnsupportedType(t *testing.T) {
	unsupportedData := TOMLTable{
		"title": []TOMLValue{1, 2, 3}, // Unsupported type for "title" field
	}

	var config TestConfig
	err := Unmarshal(unsupportedData, &config)
	if err == nil {
		t.Fatalf("Expected error when unmarshalling unsupported type, but got none")
	}

	expectedErr := "failed to set field title: expected string, got []config.TOMLValue"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %v, got: %v", expectedErr, err)
	}
}
