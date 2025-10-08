package core

import (
	"brains/internal/dag"
	"context"

	"github.com/pterm/pterm"
)

func (c *CodeData) SetResearchData(url, data string) {
	if c.ResearchData == nil {
		c.ResearchData = make(map[string]string)
	}
	c.ResearchData[url] = data
}

func (a *CodeData) generateCodeFunction(coreConfig *CoreConfig, req *LLMRequest) codeDataDAGFunction {
	additionalContext := ""
	for url, data := range a.ResearchData {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	return func(inputs map[string]string) (string, error) {
		coreConfig.Code(
			additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+req.Prompt,
			req.PersonaInstructions,
			req.ModelID,
			req.Glob,
		)

		return "", nil
	}
}

func (c *CoreConfig) CodeFlow(ctx context.Context, llmRequest *LLMRequest) error {
	codeData := &CodeData{
		ResearchData: make(map[string]string),
	}

	codeDAG, err := dag.NewDAG[string, *CodeData]("_code")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		return err
	}

	researchVertex := &dag.Vertex[string, *CodeData]{
		Name: "research",
		DAG:  codeDAG,
		Run:  generateResearchRun(c, ctx, llmRequest, codeData),
	}
	_ = codeDAG.AddVertex(researchVertex)

	codeVertex := &dag.Vertex[string, *CodeData]{
		Name: "ask",
		DAG:  codeDAG,
		Run:  codeData.generateCodeFunction(c, llmRequest),
	}
	_ = codeDAG.AddVertex(codeVertex)

	codeDAG.Connect(researchVertex.Name, codeVertex.Name)

	pterm.Success.Println("codeDAG beginning execution, planned flow printed")
	codeDAG.Visualize()

	if _, err = codeDAG.Run(); err != nil {
		pterm.Error.Printf("Failed to run DAG: %v\n", err)
		return err
	}
	pterm.Success.Println("codeDAG completed in execution successfully")

	return nil
}
