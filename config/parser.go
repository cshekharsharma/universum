package config

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type IniFile struct {
	Sections map[string]map[string]string
}

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

		ini.Sections[section][key] = value
	}

	return ini, nil
}
