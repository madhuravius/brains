package main

import (
	"context"
	"os"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/core"
)

func main() {
	const (
		glob    = "cmd/debug/commands/ask/main.go"
		modelID = "openai.gpt-oss-120b-1:0"
		prompt  = "as per best practices from https://google.github.io/styleguide/go/best-practices.html and suggest improvements to this codebase"
	)

	ctx := context.Background()
	brainsConfig, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	awsConfig := aws.NewAWSConfig(brainsConfig.AWSRegion)
	awsConfig.SetLogger(brainsConfig)
	if !awsConfig.SetAndValidateCredentials() {
		pterm.Error.Println("unable to validate credentials")
		os.Exit(1)
	}

	coreConfig := core.NewCoreConfig(awsConfig)
	coreConfig.SetLogger(brainsConfig)

	personaInstructions := brainsConfig.GetPersonaInstructions("dev")

	req := &core.LLMRequest{
		Glob:                glob,
		ModelID:             modelID,
		PersonaInstructions: personaInstructions,
		Prompt:              prompt,
	}

	coreConfig.AskFlow(ctx, req)
}
