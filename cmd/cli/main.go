package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"brains/internal/aws"
)

func main() {
	a := aws.NewAWSConfig()
	app := &cli.App{
		Name:  "brains",
		Usage: "a simple LLM wrapper using AWS Bedrock",
		Commands: []*cli.Command{
			{
				Name:  "health",
				Usage: "verify functionality and connections",
				Action: func(c *cli.Context) error {
					pterm.Info.Println("health checks starting")
					if !a.SetAndValidateCredentials() {
						pterm.Error.Println("unable to validate credentials")
						os.Exit(1)
					}
					if !a.ValidateBedrockConfiguration() {
						pterm.Error.Println("unable to access bedrock")
						os.Exit(1)
					}
					pterm.Success.Println("Health check complete")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
