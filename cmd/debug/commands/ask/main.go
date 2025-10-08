package main

import (
	"context"
	"os"

	"github.com/pterm/pterm"

	"brains/internal/aws"
	"brains/internal/config"
	"brains/internal/core"
	"brains/internal/dag"
	"brains/internal/tools/browser"
)

const (
	glob    = "cmd/debug/commands/ask/main.go"
	modelID = "openai.gpt-oss-120b-1:0"
	prompt  = "as per best practices from https://google.github.io/styleguide/go/best-practices.html and suggest improvements to this codebase"
)

type ResearchData map[string]string
type AskData struct {
	Research ResearchData
}

type askDataDAGFunction func(inputs map[string]string) (string, error)

func (a *AskData) generateResearchRun(coreConfig *core.CoreConfig, ctx context.Context) askDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		researchActions := coreConfig.Research(prompt, modelID, glob)
		for _, url := range researchActions.UrlsRecommended {
			data, err := browser.FetchWebContext(ctx, url)
			if err != nil {
				pterm.Error.Printf("Failed to load url: %v\n", err)
				os.Exit(1)
			}
			a.Research[url] = data
		}
		return "", nil
	}
}

func (a *AskData) generateAskFunction(brainsConfig *config.BrainsConfig, coreConfig *core.CoreConfig) askDataDAGFunction {
	additionalContext := ""
	for url, data := range a.Research {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	return func(inputs map[string]string) (string, error) {
		coreConfig.Ask(
			additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+prompt,
			brainsConfig.GetPersonaInstructions("dev"),
			modelID,
			glob,
		)

		return "", nil
	}
}

func main() {
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

	askData := &AskData{
		Research: make(map[string]string),
	}

	askDAG, err := dag.NewDAG[string, *AskData]("_ask")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		os.Exit(1)
	}

	researchVertex := &dag.Vertex[string, *AskData]{
		Name: "research",
		DAG:  askDAG,
		Run:  askData.generateResearchRun(coreConfig, ctx),
	}
	_ = askDAG.AddVertex(researchVertex)

	askVertex := &dag.Vertex[string, *AskData]{
		Name: "ask",
		DAG:  askDAG,
		Run:  askData.generateAskFunction(brainsConfig, coreConfig),
	}
	_ = askDAG.AddVertex(askVertex)

	askDAG.Connect(researchVertex.Name, askVertex.Name)
	askDAG.Visualize()

	_, _ = askDAG.Run()
	pterm.Success.Println("askDAG completed in execution successfully")
}
