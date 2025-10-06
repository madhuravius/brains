package core

import (
	awsConfig "brains/internal/aws"
	brainsConfig "brains/internal/config"
)

type CoreConfig struct {
	awsConfig *awsConfig.AWSConfig
	logger    brainsConfig.SimpleLogger
}
