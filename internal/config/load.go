package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultPath = "/config/config.yaml"

// Load reads the config file at path. If the file does not exist an empty
// Config is returned so the user can bootstrap via the TUI.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
