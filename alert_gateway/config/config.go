package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App map[string]any
	Log map[string]any
}

func NewConfig() *Config {
	return &Config{
		App: make(map[string]any),
		Log: make(map[string]any),
	}
}

func LoadConfig(configPath string) (*Config, error) {
	config := NewConfig()
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error loading config file file :%w", err)
	}
	return config, nil
}
