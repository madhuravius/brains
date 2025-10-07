package main

import (
	"context"
	"encoding/json"

	"brains/internal/aws"
	"brains/internal/config"
	"brains/internal/dag"
	"brains/internal/tools/browser"
	"brains/internal/tools/file_system"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Fatal.Printf("load config: %v", err)
	}

	awsCfg := aws.NewAWSConfig(cfg.AWSRegion)
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

	fs, err := file_system.NewFileSystemConfig()
	if err != nil {
		pterm.Fatal.Printfln("file_system.NewFileSystemConfig: %v", err)
	}

	fsData, err := fs.SetContextFromGlob("README.md")
	if err != nil {
		pterm.Fatal.Printfln("file_system.SetContextFromGlob: %v", err)
	}
	pterm.Info.Printfln("data from glob gather: %s", fsData)

	htmlData, err := browser.FetchWebContext(context.Background(), "https://github.com/madhuravius")
	if err != nil {
		pterm.Fatal.Printfln("browser.FetchWebContext: %v", err)
	}
	pterm.Info.Printfln("data from web gather: %s", htmlData)

	d, err := dag.NewDAG[int]("_dag")
	if err != nil {
		pterm.Fatal.Printfln("dag.NewDAG: %v", err)
	}

	v1 := &dag.Vertex[int]{Name: "a"}
	v2 := &dag.Vertex[int]{Name: "b"}
	v3 := &dag.Vertex[int]{Name: "c"}

	_ = d.AddVertex(v1)
	_ = d.AddVertex(v2)
	_ = d.AddVertex(v3)

	d.Connect("root", v1.Name)
	d.Connect(v1.Name, v2.Name)
	d.Connect(v2.Name, v3.Name)
	d.Connect(v1.Name, v3.Name)
	d.Visualize()
}
