package main

import (
	"context"
	"os"

	"atomicgo.dev/keyboard"
	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/core"
)

func main() {
	const (
		glob    = "README.md"
		modelID = "openai.gpt-oss-120b-1:0"
		prompt  = "make suggestions to the README.md in root based on this example: https://raw.githubusercontent.com/RichardLitt/standard-readme/refs/heads/main/example-readmes/minimal-readme.md"
	)

	ctx := context.Background()
	brainsConfig, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("failed to load configuration: %v\n", err)
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

	go func() {
		_ = keyboard.SimulateKeyPress('N')
	}()

	if err := coreConfig.CodeFlow(ctx, req); err != nil {
		pterm.Error.Printf("error on coreConfig.CodeFlow: %v\n", err)
		os.Exit(1)
	}
}
