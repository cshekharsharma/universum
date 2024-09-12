// Package config provides functionality for reading and parsing INI-style configuration
// files. It organizes configuration data into sections and key-value pairs, allowing
// structured access to the data.

package config

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// IniFile represents an INI file structure, where configuration data is organized
// into sections and key-value pairs.
//
// Fields:
//   - Sections: A map where the key is the section name (string) and the value is
//     another map representing the key-value pairs within that section.
type IniFile struct {
	Sections map[string]map[string]string
}

// NewIniFile loads and parses an INI file from the given filename. It returns an
// IniFile object containing the parsed sections and key-value pairs.
//
// Parameters:
//   - filename: The path to the INI file to be loaded.
//
// Returns:
//   - *IniFile: A pointer to the newly created IniFile object.
//   - error: If the file cannot be opened or read, an error is returned. This can
//     also occur if the file is improperly formatted.
//
// The INI file format supports sections, which are enclosed in square brackets ([]),
// and key-value pairs that are separated by an equals sign (=). Comments are denoted
// by lines starting with a semicolon (;).
//
// Example INI file format:
//
//	[SectionName]
//	key1 = value1
//	key2 = value2
//
// Empty lines and comment lines (starting with ;) are ignored.
func NewIniFile(filename string) (*IniFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

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

		// Add key-value pair to the current section
		ini.Sections[section][key] = value
	}

	return ini, nil
}
