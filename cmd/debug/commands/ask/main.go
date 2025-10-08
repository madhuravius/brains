package main

import (
	"context"
	"os"

	"brains/internal/aws"
	"brains/internal/config"
	"brains/internal/core"
	"brains/internal/tools/browser"

	"github.com/pterm/pterm"
)

const (
	glob    = "cmd/debug/commands/ask/main.go"
	modelID = "openai.gpt-oss-120b-1:0"
	prompt  = "as per best practices from https://google.github.io/styleguide/go/best-practices.html and suggest improvements to this codebase"
)

func main() {
	ctx := context.Background()
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

	researchActions := coreConfig.Research(prompt, modelID, glob)
	additionalContext := ""
	for _, url := range researchActions.UrlsRecommended {
		data, err := browser.FetchWebContext(ctx, url)
		if err != nil {
			pterm.Error.Printf("Failed to load url: %v\n", err)
			os.Exit(1)
		}
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}

	coreConfig.Ask(
		additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+prompt,
		cfg.GetPersonaInstructions("dev"),
		modelID,
		glob,
	)
}
