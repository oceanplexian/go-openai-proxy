package internal

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	ErrCouldNotReadFile      = errors.New("could not read file")
	ErrCouldNotUnmarshalYAML = errors.New("could not unmarshal YAML")
)

func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCouldNotReadFile, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCouldNotUnmarshalYAML, err)
	}

	return &cfg, nil
}
