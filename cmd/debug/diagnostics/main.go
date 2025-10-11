package main

import (
	"encoding/json"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/config"

	"github.com/pterm/pterm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Fatal.Printf("load config: %v", err)
	}

	awsCfg := aws.NewAWSConfig(cfg.GetConfig().AWSRegion)
	if !awsCfg.SetAndValidateCredentials() {
		pterm.Fatal.Printf("invalid AWS credentials")
	}

	pterm.Info.Printf("Region: %s\nModel: %s\n",
		cfg.GetConfig().AWSRegion, cfg.GetConfig().Model)

	data := awsCfg.DescribeModel(cfg.GetConfig().Model)
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
