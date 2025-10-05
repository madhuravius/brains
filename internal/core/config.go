package core

import (
	"brains/internal/aws"
	brainsConfig "brains/internal/config"
)

func NewCoreConfig(awsConfig *aws.AWSConfig) *CoreConfig {
	return &CoreConfig{
		awsConfig: awsConfig,
	}
}
func (c *CoreConfig) GetAWSConfig() *aws.AWSConfig          { return c.awsConfig }
func (c *CoreConfig) SetAWSConfig(a *aws.AWSConfig)         { c.awsConfig = a }
func (c *CoreConfig) SetLogger(l brainsConfig.SimpleLogger) { c.logger = l }
