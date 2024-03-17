package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	OutputPath string       `json:"outputPath"`
	Configs    []EnumConfig `json:"enums"`
}

// Validate validates config.
func (c Config) Validate() error {
	if c.OutputPath == "" {
		return errors.New("output path can not be empty")
	}

	if len(c.Configs) == 0 {
		return errors.New("enum config can not be empty")
	}

	for _, ec := range c.Configs {
		err := ec.validate()
		if err != nil {
			return fmt.Errorf("error validate enum config: %w", err)
		}
	}

	return nil
}

type EnumConfig struct {
	Package string   `json:"package"`
	Type    string   `json:"type"`
	Enums   []string `json:"values"`
}

func ReadConfig(path string) (Config, error) {
	// read config file from the json file and return the config
	file, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

// validate validates enum config.
func (ec EnumConfig) validate() error {
	if ec.Package == "" {
		return errors.New("package can not be empty")
	}
	if ec.Type == "" {
		return errors.New("type can not be empty")
	}
	if len(ec.Enums) == 0 {
		return errors.New("enum can not be empty")
	}
	return nil
}
