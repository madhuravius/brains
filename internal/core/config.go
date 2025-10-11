package core

import (
	"os"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	brainsConfig "github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/tools/file_system"
)

func NewCoreConfig(awsConfig *aws.AWSConfig) CoreImpl {
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
