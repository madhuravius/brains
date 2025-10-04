package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"brains/internal/aws"
	"brains/internal/config"
)

type CLIConfig struct {
	awsConfig    *aws.AWSConfig
	brainsConfig *config.BrainsConfig
	persona      string
	glob         string
}

func generateCommonFlags(cliConfig *CLIConfig) []cli.Flag {
	return []cli.Flag{
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
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	awsConfig := aws.NewAWSConfig(cfg.Model, cfg.AWSRegion)
	awsConfig.SetLogger(cfg)

	cliConfig := CLIConfig{
		awsConfig:    awsConfig,
		brainsConfig: cfg,
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
					pterm.Success.Println("health check complete")
					return nil
				},
			},
			{
				Name:  "ask",
				Usage: "send a prompt to the Bedrock model and display the response",
				Flags: generateCommonFlags(&cliConfig),
				Action: func(c *cli.Context) error {
					prompt := c.Args().Get(0)
					if prompt == "" {
						textInput := pterm.DefaultInteractiveTextInput.WithMultiLine()
						prompt, _ = textInput.Show()
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

					pterm.Success.Println("question answered")
					return nil
				},
			},
			{
				Name:  "code",
				Usage: "send a prompt to the Bedrock model and execute coding actions",
				Flags: generateCommonFlags(&cliConfig),
				Action: func(c *cli.Context) error {
					prompt := c.Args().Get(0)
					if prompt == "" {
						textInput := pterm.DefaultInteractiveTextInput.WithMultiLine()
						prompt, _ = textInput.Show()
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
					if !cliConfig.awsConfig.Code(prompt, personaInstructions, addedContext) {
						os.Exit(0)
					}

					pterm.Success.Println("code execution complete")
					return nil
				},
			},
			{
				Name:  "log",
				Usage: "print all logs",
				Action: func(c *cli.Context) error {
					cliConfig.brainsConfig.PrintLogs()
					pterm.Success.Println("logs printed.")
					return nil
				},
			},
			{
				Name:  "reset",
				Usage: "clear all logs",
				Action: func(c *cli.Context) error {
					if err := cliConfig.brainsConfig.Reset(); err != nil {
						pterm.Error.Printfln("reset failed: %v", err)
						return err
					}
					pterm.Success.Println("logs cleared.")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
