package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type logger struct {
	enabled bool
	file    *os.File
	mu      sync.Mutex
	mem     []string
}

type SimpleLogger interface {
	LogMessage(string)
	GetLogContext() string
}

type BrainsConfig struct {
	LoggingEnabled bool              `yaml:"logging_enabled"`
	AWSRegion      string            `yaml:"aws_region"`
	Model          string            `yaml:"model"`
	Personas       map[string]string `yaml:"personas"`

	logger logger `yaml:"-"`
}

var DefaultConfig = BrainsConfig{
	LoggingEnabled: true,
	AWSRegion:      "us-west-2",
	Model:          "openai.gpt-oss-120b-1:0",
	Personas:       map[string]string{},
}

func LoadConfig() (*BrainsConfig, error) {
	paths := []string{}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".brains.yml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".brains.yml"))
	}
	var cfgPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			cfgPath = p
			break
		}
	}
	if cfgPath == "" {
		target := paths[0]
		data, _ := yaml.Marshal(&DefaultConfig)
		if err := os.WriteFile(target, data, 0o600); err != nil {
			return nil, err
		}
		return &DefaultConfig, nil
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg BrainsConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = DefaultConfig.AWSRegion
	}
	if cfg.Model == "" {
		cfg.Model = DefaultConfig.Model
	}

	if err := cfg.initLogger(cfg.LoggingEnabled); err != nil {
		return nil, err
	}
	return &cfg, nil
}
