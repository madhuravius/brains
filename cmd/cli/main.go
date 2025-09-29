package main

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"brains/internal/aws"
	"brains/internal/config"
)

type CLIConfig struct {
	awsConfig    aws.AWSConfig
	brainsConfig config.BrainsConfig
	persona      string
	glob         string
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	cliConfig := CLIConfig{
		awsConfig:    *aws.NewAWSConfig(cfg.Model, cfg.AWSRegion),
		brainsConfig: *cfg,
	}

	app := &cli.App{
		Name:  "brains",
		Usage: "a simple LLM wrapper using AWS Bedrock",
		Commands: []*cli.Command{
			{
				Name:  "health",
				Usage: "verify functionality and connections",
				Action: func(c *cli.Context) error {
					pterm.Info.Println("health checks starting")
					if !cliConfig.awsConfig.SetAndValidateCredentials() {
						pterm.Error.Println("unable to validate credentials")
						os.Exit(1)
					}
					if !cliConfig.awsConfig.ValidateBedrockConfiguration() {
						pterm.Error.Println("unable to access bedrock")
						os.Exit(1)
					}
					pterm.Success.Println("Health check complete")
					return nil
				},
			},
			{
				Name:  "ask",
				Usage: "send a prompt to the Bedrock model and display the response",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Aliases:     []string{"p"},
						Name:        "persona",
						Value:       "default",
						Usage:       "Supply a persona with a corresponding value in \".brains.yml\" to use as part of the prompt.",
						Destination: &cliConfig.persona,
					},
					&cli.StringFlag{
						Aliases:     []string{"a"},
						Name:        "add",
						Value:       "",
						Usage:       "Supply a glob pattern to add to the context",
						Destination: &cliConfig.glob,
					},
				},
				Action: func(c *cli.Context) error {
					prompt := c.Args().Get(0)
					if prompt == "" {
						return fmt.Errorf("prompt argument required")
					}
					if !cliConfig.awsConfig.SetAndValidateCredentials() {
						pterm.Error.Println("unable to validate credentials")
						os.Exit(1)
					}
					personaInstructions := cliConfig.brainsConfig.GetPersonaInstructions(cliConfig.persona)
					addedContext := ""
					if cliConfig.glob != "" {
						var err error
						addedContext, err = cliConfig.brainsConfig.SetContextFromGlob(cliConfig.glob)
						if err != nil {
							pterm.Error.Printfln("failed to read glob pattern for context: %v", err)
							os.Exit(1)
						}

					}
					if !cliConfig.awsConfig.Ask(prompt, personaInstructions, addedContext) {
						pterm.Error.Println("failed to get response from Bedrock")
						os.Exit(1)
					}

					pterm.Success.Println("Question answered.")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
