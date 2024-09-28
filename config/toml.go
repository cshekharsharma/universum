package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TOMLValue interface{}
type TOMLTable map[string]TOMLValue

// Parser represents the TOML parser
type Parser struct {
	data      TOMLTable
	currTable TOMLTable
	section   string
}

// NewParser creates a new TOML parser
func NewParser() *Parser {
	return &Parser{
		data: make(TOMLTable),
	}
}

// ParseLine parses a single line of TOML input
func (p *Parser) ParseLine(line string) error {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		// Ignore empty lines and comments
		return nil
	}

	// Handle sections (e.g., [section])
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		sectionName := strings.Trim(line, "[]")
		// Handle nested sections by splitting on dots
		nestedSections := strings.Split(sectionName, ".")
		p.currTable = p.ensureNestedTables(nestedSections)
		p.section = sectionName
		return nil
	}

	// Handle key-value pairs
	if strings.Contains(line, "=") {
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse value
		parsedValue, err := p.parseValue(value)
		if err != nil {
			return err
		}

		// Store value in the current section/table
		if p.currTable != nil {
			p.currTable[key] = parsedValue
		} else {
			p.data[key] = parsedValue
		}
	}

	return nil
}

// ensureNestedTables ensures that nested tables exist and returns the final table
func (p *Parser) ensureNestedTables(sections []string) TOMLTable {
	currTable := p.data

	for _, section := range sections {
		if _, exists := currTable[section]; !exists {
			currTable[section] = make(TOMLTable)
		}

		// Move to the next nested table
		currTable = currTable[section].(TOMLTable)
	}

	return currTable
}

// parseValue parses the value into its correct type
func (p *Parser) parseValue(value string) (TOMLValue, error) {
	// Handle boolean
	if value == "true" || value == "false" {
		return strconv.ParseBool(value)
	}

	// Handle integer
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue, nil
	}

	// Handle float
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue, nil
	}

	// Handle string (with quotes)
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return strings.Trim(value, `"`), nil
	}

	// Handle array
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return p.parseArray(value)
	}

	return nil, fmt.Errorf("unknown value type: %s", value)
}

// parseArray parses an array from the input
func (p *Parser) parseArray(value string) ([]TOMLValue, error) {
	array := []TOMLValue{}
	arrayContent := strings.Trim(value, "[]")
	elements := strings.Split(arrayContent, ",")
	for _, element := range elements {
		trimmedElement := strings.TrimSpace(element)
		parsedValue, err := p.parseValue(trimmedElement)
		if err != nil {
			return nil, err
		}
		array = append(array, parsedValue)
	}
	return array, nil
}

// Parse parses the TOML input file line by line
func (p *Parser) Parse(file *os.File) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		err := p.ParseLine(line)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
