package config

import (
	"os"
	"sync"
)

type logger struct {
	enabled bool
	file    *os.File
	mu      sync.Mutex
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
	DefaultContext string            `yaml:"default_context"`
	DefaultPersona string            `yaml:"default_persona"`
	logger         logger            `yaml:"-"`
}

type BrainsConfigImpl interface {
	GetConfig() *BrainsConfig
	GetPersonaInstructions(persona string) string
}
