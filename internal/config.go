package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config file read failed: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return nil, fmt.Errorf("yaml parse failed: %w", err)
	}

	return &cfg, nil
}
