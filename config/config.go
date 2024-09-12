// Package config provides functionality for loading and accessing application
// configuration values from an INI file. It supports retrieving string, integer,
// float, and boolean values from specific sections within the configuration file.

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"universum/utils"
)

// CONFIG holds the parsed configuration file as an instance of IniFile.
// The configuration file is loaded when the Load function is called.
var CONFIG *IniFile

// INVALID_NUMERIC_VALUE is a constant representing an invalid numeric value
// returned when a requested configuration key is not found.
const INVALID_NUMERIC_VALUE = -99999999

// INVALID_EPOCH_VALUE is a constant representing an invalid epoch value,
// used when a date or time-related key is not found or improperly formatted.
const INVALID_EPOCH_VALUE = 0

// Load loads the configuration from the provided file path into the CONFIG variable.
//
// Parameters:
//   - filepath: A string representing the full path to the configuration file.
//
// Returns:
//   - error: If the file cannot be loaded, an error is returned. If the path is
//     invalid or empty, it returns an error indicating the issue. Otherwise, it
//     returns nil.
func Load(filepath string) error {
	var err error

	if len(filepath) < 1 {
		return errors.New("error loading config file: Invalid/empty config path provided")
	}

	CONFIG, err = NewIniFile(filepath)

	if err != nil {
		return fmt.Errorf("error loading config file: %v", err)
	}

	return nil
}

// Get retrieves a configuration value as an interface{} from the specified section and key.
//
// Parameters:
//   - key: The key of the configuration value to retrieve.
//   - section: The section in which the key resides.
//
// Returns:
//   - any: The value of the configuration key. If the section or key is not found, it returns nil.
func Get(key string, section string) any {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return value
		}
	}
	return nil
}

// GetInt64 retrieves an integer value from the configuration file.
//
// Parameters:
//   - key: The key of the configuration value to retrieve.
//   - section: The section in which the key resides.
//
// Returns:
//   - int64: The integer value for the given key. If the key is not found or
//     conversion fails, it returns INVALID_NUMERIC_VALUE.
//   - error: An error if the key is not found or conversion fails.
func GetInt64(key string, section string) (int64, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseInt(value, 10, 64)
		}
	}
	return INVALID_NUMERIC_VALUE, errors.New("no config found for provided key-section pair")
}

// GetFloat64 retrieves a floating-point value from the configuration file.
//
// Parameters:
//   - key: The key of the configuration value to retrieve.
//   - section: The section in which the key resides.
//
// Returns:
//   - float64: The floating-point value for the given key. If the key is not found
//     or conversion fails, it returns INVALID_NUMERIC_VALUE.
//   - error: An error if the key is not found or conversion fails.
func GetFloat64(key string, section string) (float64, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseFloat(value, 64)
		}
	}
	return INVALID_NUMERIC_VALUE, errors.New("no config found for provided key-section pair")
}

// GetString retrieves a string value from the configuration file.
//
// Parameters:
//   - key: The key of the configuration value to retrieve.
//   - section: The section in which the key resides.
//
// Returns:
//   - string: The string value for the given key. If the key is not found, it returns an empty string.
//   - error: An error if the key is not found.
func GetString(key string, section string) (string, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return value, nil
		}
	}
	return "", errors.New("no config found for provided key-section pair")
}

// GetBool retrieves a boolean value from the configuration file.
//
// Parameters:
//   - key: The key of the configuration value to retrieve.
//   - section: The section in which the key resides.
//
// Returns:
//   - bool: The boolean value for the given key. If the key is not found or
//     conversion fails, it returns false.
//   - error: An error if the key is not found or conversion fails.
func GetBool(key string, section string) (bool, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseBool(value)
		}
	}
	return false, errors.New("no config found for provided key-section pair")
}

// GetDefaultConfigPath retrieves the default configuration path based on the operating system.
// It checks the environment variable "config", and if not set, returns platform-specific paths.
//
// Returns:
//   - string: The default configuration path based on the environment variable or operating system.
func GetDefaultConfigPath() string {
	configpath := os.Getenv("config")

	if len(configpath) > 1 {
		return configpath
	}

	if utils.IsDarwin() {
		return fmt.Sprintf("/Library/Application Support/%s/%s", APP_CODE_NAME, DEFAULT_CONFIG_NAME)
	}

	// Assuming if it's not Darwin, then it's Linux.
	return fmt.Sprintf("/etc/%s/%s", APP_CODE_NAME, DEFAULT_CONFIG_NAME)
}
