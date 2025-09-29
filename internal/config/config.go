package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BrainsConfig struct {
	AWSRegion string            `yaml:"aws_region"`
	Model     string            `yaml:"model"`
	Personas  map[string]string `yaml:"personas"`
}

var DefaultConfig = BrainsConfig{
	AWSRegion: "us-west-2",
	Model:     "openai.gpt-oss-120b-1:0",
	Personas:  map[string]string{},
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
		_ = os.WriteFile(target, data, 0o644)
		c := DefaultConfig
		return &c, nil
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
	return &cfg, nil
}
