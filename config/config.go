package config

import (
	"errors"
	"fmt"
	"os"
)

var Store *Config
var ConfigFilePath string

func Init(filepath string) error {
	var err error
	Store = GetSkeleton()

	if len(filepath) < 1 {
		return errors.New("error loading config file: Invalid/empty config path provided")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("could not read config file: %v", err)
	}

	defer file.Close()
	ConfigFilePath = filepath

	parser := NewParser()
	err = parser.Parse(file)

	if err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	err = Unmarshal(parser.data, Store)
	if err != nil {
		return fmt.Errorf("error unmarshalling config: %v", err)
	}

	validator := NewConfigValidator()
	_, err = validator.Validate(Store)

	if err != nil {
		return fmt.Errorf("error validating config: %v", err)
	}

	return nil
}

func GetSkeleton() *Config {
	return &Config{
		Server:  &Server{},
		Cluster: &Cluster{},
		Storage: &Storage{
			LSM:    &LSM{},
			Memory: &Memory{},
		},
		Eviction: &Eviction{},
		Logging:  &Logging{},
		Auth:     &Auth{},
	}
}

func PopulateDefaultConfig() {
	if Store == nil {
		Store = GetSkeleton()
		v := NewConfigValidator()
		_, _ = v.Validate(Store)
	}
}
