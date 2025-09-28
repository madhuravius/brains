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
	a aws.AWSConfig
	b config.BrainsConfig
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	cliConfig := CLIConfig{
		a: *aws.NewAWSConfig(cfg.Model),
		b: *cfg,
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
					if !cliConfig.a.SetAndValidateCredentials() {
						pterm.Error.Println("unable to validate credentials")
						os.Exit(1)
					}
					if !cliConfig.a.ValidateBedrockConfiguration() {
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
				Action: func(c *cli.Context) error {
					prompt := c.Args().Get(0)
					if prompt == "" {
						return fmt.Errorf("prompt argument required")
					}
					if !cliConfig.a.SetAndValidateCredentials() {
						pterm.Error.Println("unable to validate credentials")
						os.Exit(1)
					}
					if !cliConfig.a.Ask(prompt) {
						pterm.Error.Println("failed to get response from Bedrock")
						os.Exit(1)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
