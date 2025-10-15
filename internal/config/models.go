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

type ContextConfig struct {
	SendLogs      bool `yaml:"send_logs"`
	SummarizeLogs bool `yaml:"summarize_logs"`
	SendAllTags   bool `yaml:"send_all_tags"`
	SendFileList  bool `yaml:"send_file_list"`
}

type BrainsConfig struct {
	LoggingEnabled bool              `yaml:"logging_enabled"`
	AWSRegion      string            `yaml:"aws_region"`
	Model          string            `yaml:"model"`
	Personas       map[string]string `yaml:"personas"`
	DefaultContext string            `yaml:"default_context"`
	DefaultPersona string            `yaml:"default_persona"`
	PreCommands    []string          `yaml:"pre_commands"`
	ContextConfig  ContextConfig     `yaml:"context_config"`

	logger logger `yaml:"-"`
}

type BrainsConfigImpl interface {
	GetConfig() *BrainsConfig
	GetPersonaInstructions(persona string) string
}
