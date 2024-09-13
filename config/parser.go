// Package config provides functionality for reading and parsing INI-style configuration files.
// The configuration data is organized into sections and key-value pairs.
package config

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// IniFile represents a parsed INI file, with sections containing key-value pairs.
type IniFile struct {
	Sections map[string]map[string]string
}

// ReadConfig reads and parses an INI file from the provided filename, returning an IniFile
// object containing sections and key-value pairs.
//
// Returns an error if the file cannot be opened or if the format is invalid.
func ReadConfig(filename string) (*IniFile, error) {
	configPath = filename

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	return parseIniFile(reader)
}

// parseIniFile parses the content of an INI file from the given reader.
// It handles unquoted, single-quoted, and double-quoted values.
func parseIniFile(reader *bufio.Reader) (*IniFile, error) {
	ini := &IniFile{
		Sections: make(map[string]map[string]string),
	}

	var section string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == ';' {
			continue
		}

		if line[0] == '[' && line[len(line)-1] == ']' {
			section = line[1 : len(line)-1]
			ini.Sections[section] = make(map[string]string)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		ini.Sections[section][key] = value
	}

	return ini, nil
}
