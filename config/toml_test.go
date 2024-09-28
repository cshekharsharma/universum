package config

import (
	"os"
	"testing"
)

func createFile(t *testing.T, content string) *os.File {
	file, err := os.CreateTemp("", "example.toml")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Fatalf("Error resetting file pointer: %v", err)
	}
	return file
}

// Test for parsing basic key-value pairs
func TestParseKeyValue(t *testing.T) {
	content := `
title = "TOML Example"
enabled = true
count = 42
pi = 3.14159
`
	file := createFile(t, content)
	defer os.Remove(file.Name())

	parser := NewParser()
	err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Error parsing TOML: %v", err)
	}

	if parser.data["title"] != "TOML Example" {
		t.Errorf("Expected title to be 'TOML Example', got %v", parser.data["title"])
	}

	if parser.data["enabled"] != true {
		t.Errorf("Expected enabled to be true, got %v", parser.data["enabled"])
	}

	if parser.data["count"] != 42 {
		t.Errorf("Expected count to be 42, got %v", parser.data["count"])
	}

	if parser.data["pi"] != 3.14159 {
		t.Errorf("Expected pi to be 3.14159, got %v", parser.data["pi"])
	}
}

func TestParseArray(t *testing.T) {
	content := `
numbers = [1, 2, 3, 4, 5]
fruits = ["apple", "banana", "cherry"]
`
	file := createFile(t, content)
	defer os.Remove(file.Name())

	parser := NewParser()
	err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Error parsing TOML: %v", err)
	}

	expectedNumbers := []TOMLValue{1, 2, 3, 4, 5}
	expectedFruits := []TOMLValue{"apple", "banana", "cherry"}

	if !equalArrays(parser.data["numbers"], expectedNumbers) {
		t.Errorf("Expected numbers to be %v, got %v", expectedNumbers, parser.data["numbers"])
	}

	if !equalArrays(parser.data["fruits"], expectedFruits) {
		t.Errorf("Expected fruits to be %v, got %v", expectedFruits, parser.data["fruits"])
	}
}

func TestParseSections(t *testing.T) {
	content := `
[database]
host = "localhost"
port = 5432
enabled = true
`
	file := createFile(t, content)
	defer os.Remove(file.Name())

	parser := NewParser()
	err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Error parsing TOML: %v", err)
	}

	section := parser.data["database"].(TOMLTable)

	if section["host"] != "localhost" {
		t.Errorf("Expected host to be 'localhost', got %v", section["host"])
	}

	if section["port"] != 5432 {
		t.Errorf("Expected port to be 5432, got %v", section["port"])
	}

	if section["enabled"] != true {
		t.Errorf("Expected enabled to be true, got %v", section["enabled"])
	}
}

func TestParseNestedTables(t *testing.T) {
	content := `
[owner]
name = "Tom"
age = 36

[owner.address]
city = "New York"
zip = "10001"
`
	file := createFile(t, content)
	defer os.Remove(file.Name())

	parser := NewParser()
	err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Error parsing TOML: %v", err)
	}

	owner := parser.data["owner"].(TOMLTable)
	address := owner["address"].(TOMLTable)

	if owner["name"] != "Tom" {
		t.Errorf("Expected name to be 'Tom', got %v", owner["name"])
	}

	if owner["age"] != 36 {
		t.Errorf("Expected age to be 36, got %v", owner["age"])
	}

	if address["city"] != "New York" {
		t.Errorf("Expected city to be 'New York', got %v", address["city"])
	}

	if address["zip"] != "10001" {
		t.Errorf("Expected zip to be '10001', got %v", address["zip"])
	}
}

func equalArrays(a, b interface{}) bool {
	arrayA, okA := a.([]TOMLValue)
	arrayB, okB := b.([]TOMLValue)
	if !okA || !okB {
		return false
	}

	if len(arrayA) != len(arrayB) {
		return false
	}

	for i := range arrayA {
		if arrayA[i] != arrayB[i] {
			return false
		}
	}

	return true
}
