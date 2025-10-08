package main

import (
	"os"

	"brains/internal/aws"
	"brains/internal/config"
	"brains/internal/core"

	"github.com/pterm/pterm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	awsConfig := aws.NewAWSConfig(cfg.AWSRegion)
	awsConfig.SetLogger(cfg)
	if !awsConfig.SetAndValidateCredentials() {
		pterm.Error.Println("unable to validate credentials")
		os.Exit(1)
	}

	coreConfig := core.NewCoreConfig(awsConfig)
	coreConfig.SetLogger(cfg)

	coreConfig.Ask(
		"go to https://google.github.io/styleguide/go/best-practices.html and suggest improvements to this codebase",
		cfg.GetPersonaInstructions("dev"),
		"openai.gpt-oss-120b-1:0",
		"cmd/debug/commands/ask/main.go",
	)
}
