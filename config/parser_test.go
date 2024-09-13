package config

import (
	"bufio"
	"strings"
	"testing"
)

func TestNewIniFile(t *testing.T) {
	iniContent := `
		[Section1]
		key1 = value1
		key2 = "double quoted value"
		key3 = 'single quoted value'
		key4 =    value with spaces   

		[Section2]
		keyA = unquotedValue
		keyB = " value with leading and trailing spaces "
		; this is a comment and should be ignored
	`

	stringReader := strings.NewReader(iniContent)
	reader := bufio.NewReader(stringReader)

	iniFile, err := parseIniFile(reader)
	if err != nil {
		t.Fatalf("Failed to parse INI file: %v", err)
	}

	expected := map[string]map[string]string{
		"Section1": {
			"key1": "value1",
			"key2": "double quoted value",
			"key3": "single quoted value",
			"key4": "value with spaces",
		},
		"Section2": {
			"keyA": "unquotedValue",
			"keyB": " value with leading and trailing spaces ",
		},
	}

	for section, keys := range expected {
		for key, expectedValue := range keys {
			actualValue := iniFile.Sections[section][key]
			if actualValue != expectedValue {
				t.Errorf("For section '%s' and key '%s', expected '%s', but got '%s'", section, key, expectedValue, actualValue)
			}
		}
	}

}
