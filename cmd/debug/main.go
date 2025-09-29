package main

import (
	"context"
	"encoding/json"

	"brains/internal/aws"
	"brains/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Fatal.Printf("load config: %v", err)
	}

	awsCfg := aws.NewAWSConfig(cfg.Model, cfg.AWSRegion)
	if !awsCfg.SetAndValidateCredentials() {
		pterm.Fatal.Printf("invalid AWS credentials")
	}

	caller, err := sts.NewFromConfig(awsCfg.GetConfig()).GetCallerIdentity(context.Background(),
		&sts.GetCallerIdentityInput{})
	if err != nil {
		pterm.Fatal.Printfln("sts identity: %v", err)
	}
	pterm.Info.Printf("Account: %s\nARN: %s\nRegion: %s\nModel: %s\n",
		*caller.Account, *caller.Arn, cfg.AWSRegion, cfg.Model)

	data := awsCfg.DescribeModel(cfg.Model)
	if err != nil {
		pterm.Fatal.Printfln("DescribeModel: %v", err)
	}
	if data == nil {
		pterm.Fatal.Printfln("Empty data model: %v", err)
	}

	rawData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		pterm.Fatal.Printfln("MarshalData: %v", err)
	}

	jsonData := make(map[string]any)
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		pterm.Fatal.Printfln("UnmarshalData: %v", err)
	}

	pterm.DefaultLogger.Info("Model details", pterm.Logger.ArgsFromMap(pterm.DefaultLogger, jsonData))
}
