package core

import (
	"brains/internal/agents/file_system"
	awsConfig "brains/internal/aws"
	brainsConfig "brains/internal/config"
)

type agentsConfig struct {
	fsAgentConfig *file_system.FileSystemConfig
}

type CoreConfig struct {
	agentsConfig *agentsConfig
	awsConfig    *awsConfig.AWSConfig
	logger       brainsConfig.SimpleLogger
}
