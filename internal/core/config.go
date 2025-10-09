package core

import (
	"os"

	"github.com/pterm/pterm"

	"brains/internal/aws"
	brainsConfig "brains/internal/config"
	"brains/internal/tools/file_system"
)

func NewCoreConfig(awsConfig *aws.AWSConfig) *CoreConfig {
	fsToolConfig, err := file_system.NewFileSystemConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load fs tool configuration: %v\n", err)
		os.Exit(1)
	}
	return &CoreConfig{
		toolsConfig: &toolsConfig{
			fsToolConfig: fsToolConfig,
		},
		awsConfig: awsConfig,
	}
}
func (c *CoreConfig) GetAWSConfig() *aws.AWSConfig          { return c.awsConfig }
func (c *CoreConfig) SetAWSConfig(a *aws.AWSConfig)         { c.awsConfig = a }
func (c *CoreConfig) SetLogger(l brainsConfig.SimpleLogger) { c.logger = l }
