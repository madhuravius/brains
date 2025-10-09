package core

import (
	"context"
	"os"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/dag"
)

func (a *AskData) SetResearchData(url, data string) {
	if a.ResearchData == nil {
		a.ResearchData = make(map[string]string)
	}
	a.ResearchData[url] = data
}

func (a *AskData) generateAskFunction(coreConfig *CoreConfig, req *LLMRequest) askDataDAGFunction {
	additionalContext := ""
	for url, data := range a.ResearchData {
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
		ResearchData: make(map[string]string),
	}

	askDAG, err := dag.NewDAG[string, *AskData]("_ask")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		os.Exit(1)
	}

	researchVertex := &dag.Vertex[string, *AskData]{
		Name: "research",
		DAG:  askDAG,
		Run:  generateResearchRun(c, ctx, llmRequest, askData),
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
