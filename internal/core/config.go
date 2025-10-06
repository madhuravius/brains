package core

import (
	"os"

	"github.com/pterm/pterm"

	"brains/internal/agents/file_system"
	"brains/internal/aws"
	brainsConfig "brains/internal/config"
)

func NewCoreConfig(awsConfig *aws.AWSConfig) *CoreConfig {
	fsAgentConfig, err := file_system.NewFileSystemConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load fs agent configuration: %v\n", err)
		os.Exit(1)
	}
	return &CoreConfig{
		agentsConfig: &agentsConfig{
			fsAgentConfig: fsAgentConfig,
		},
		awsConfig: awsConfig,
	}
}
func (c *CoreConfig) GetAWSConfig() *aws.AWSConfig          { return c.awsConfig }
func (c *CoreConfig) SetAWSConfig(a *aws.AWSConfig)         { c.awsConfig = a }
func (c *CoreConfig) SetLogger(l brainsConfig.SimpleLogger) { c.logger = l }
