package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"universum/utils"
)

var CONFIG *IniFile

const INVALID_NUMERIC_VALUE = -99999999
const INVALID_EPOCH_VALUE = 0

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

func Get(key string, section string) any {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return value
		}
	}

	return nil
}

func GetInt64(key string, section string) (int64, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseInt(value, 10, 64)
		}
	}

	return INVALID_NUMERIC_VALUE, errors.New("no config found for provided key-section pair")
}

func GetFloat64(key string, section string) (float64, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseFloat(value, 64)
		}
	}

	return INVALID_NUMERIC_VALUE, errors.New("no config found for provided key-section pair")
}

func GetString(key string, section string) (string, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return value, nil
		}
	}

	return "", errors.New("no config found for provided key-section pair")

}

func GetBool(key string, section string) (bool, error) {
	if section, ok := CONFIG.Sections[section]; ok {
		if value, ok := section[key]; ok {
			return strconv.ParseBool(value)
		}
	}

	return false, errors.New("no config found for provided key-section pair")
}

func GetDefaultConfigPath() string {
	configpath := os.Getenv("config")

	if len(configpath) > 1 {
		return configpath
	}

	if utils.IsDarwin() {
		return fmt.Sprintf("/Library/Application Support/%s/%s", APP_CODE_NAME, DEFAULT_CONFIG_NAME)
	}

	// assuming if its not darwin, then its Linux.
	return fmt.Sprintf("/etc/%s/%s", APP_CODE_NAME, DEFAULT_CONFIG_NAME)
}
