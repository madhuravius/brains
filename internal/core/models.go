package core

import (
	awsConfig "brains/internal/aws"
	brainsConfig "brains/internal/config"
	"brains/internal/tools/file_system"
)

type toolsConfig struct {
	fsToolConfig *file_system.FileSystemConfig
}

type CoreConfig struct {
	toolsConfig *toolsConfig
	awsConfig   *awsConfig.AWSConfig
	logger      brainsConfig.SimpleLogger
}
