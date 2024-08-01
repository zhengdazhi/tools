package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config 结构体定义了整个配置文件的结构
type Config struct {
	Database map[string]any
	Log      map[string]any
	App      map[string]any
}

// NewConfig 创建并初始化一个新的 Config 实例
func NewConfig() *Config {
	return &Config{
		Database: make(map[string]any),
		Log:      make(map[string]any),
		App:      make(map[string]any),
	}
}

// LoadConfig 从指定路径加载并解析 TOML 配置文件
func LoadConfig(configPath string) (*Config, error) {
	// var config Config
	config := NewConfig()

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	return config, nil
}
