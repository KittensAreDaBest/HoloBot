package config

import (
	"github.com/go-yaml/yaml"
	"os"
	"path/filepath"
)

type Configuration struct {
	Token    string `yaml:"token"`
	Timezone string `yaml:"timezone"`
}

func ParseConfiguration(path string) (Configuration, error) {
	var cfg Configuration

	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
