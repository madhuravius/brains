package core

import (
	"context"
	"os"

	"github.com/pterm/pterm"

	"brains/internal/dag"
	"brains/internal/tools/browser"
)

func (a *AskData) generateResearchRun(coreConfig *CoreConfig, ctx context.Context, req *LLMRequest) askDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		researchActions := coreConfig.Research(req.Prompt, req.ModelID, req.Glob)
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

func (a *AskData) generateAskFunction(coreConfig *CoreConfig, req *LLMRequest) askDataDAGFunction {
	additionalContext := ""
	for url, data := range a.Research {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	return func(inputs map[string]string) (string, error) {
		coreConfig.Ask(
			additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+req.Prompt,
			req.PersonaInstructions,
			req.ModelID,
			req.Glob,
		)

		return "", nil
	}
}

func (c *CoreConfig) AskFlow(ctx context.Context, llmRequest *LLMRequest) {
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
		Run:  askData.generateResearchRun(c, ctx, llmRequest),
	}
	_ = askDAG.AddVertex(researchVertex)

	askVertex := &dag.Vertex[string, *AskData]{
		Name: "ask",
		DAG:  askDAG,
		Run:  askData.generateAskFunction(c, llmRequest),
	}
	_ = askDAG.AddVertex(askVertex)

	askDAG.Connect(researchVertex.Name, askVertex.Name)
	pterm.Success.Println("askDAG beginning execution, planned flow printed")
	askDAG.Visualize()

	if _, err = askDAG.Run(); err != nil {
		pterm.Error.Printf("Failed to run DAG: %v\n", err)
		os.Exit(1)
	}
	pterm.Success.Println("askDAG completed in execution successfully")
}
