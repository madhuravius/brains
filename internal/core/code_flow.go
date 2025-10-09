package core

import (
	"github.com/madhuravius/brains/internal/dag"
	"context"
	"fmt"

	"github.com/pterm/pterm"
)

func (c *CodeData) SetResearchData(url, data string) {
	if c.ResearchData == nil {
		c.ResearchData = make(map[string]string)
	}
	c.ResearchData[url] = data
}

func (c *CodeData) generateDetermineCodeChangesFunction(coreConfig *CoreConfig, req *LLMRequest) codeDataDAGFunction {
	additionalContext := ""
	for url, data := range c.ResearchData {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	return func(inputs map[string]string) (string, error) {
		c.CodeModelResponse = coreConfig.DetermineCodeChanges(
			additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+req.Prompt,
			req.PersonaInstructions,
			req.ModelID,
			req.Glob,
		)

		return "", nil
	}
}

func (c *CodeData) generateExecuteCodeEditsFunction(coreConfig *CoreConfig) codeDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		if !coreConfig.ExecuteEditCode(c.CodeModelResponse) {
			return "", fmt.Errorf("error in generateExecuteCodeEditsFunction, unable to execute edits")
		}
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

	determineCodeChangesVertex := &dag.Vertex[string, *CodeData]{
		Name: "determine_code_changes",
		DAG:  codeDAG,
		Run:  codeData.generateDetermineCodeChangesFunction(c, llmRequest),
	}
	_ = codeDAG.AddVertex(determineCodeChangesVertex)

	executeCodeEditsVertex := &dag.Vertex[string, *CodeData]{
		Name: "execute_code_edits",
		DAG:  codeDAG,
		Run:  codeData.generateExecuteCodeEditsFunction(c),
	}
	_ = codeDAG.AddVertex(executeCodeEditsVertex)

	codeDAG.Connect(researchVertex.Name, determineCodeChangesVertex.Name)
	codeDAG.Connect(determineCodeChangesVertex.Name, executeCodeEditsVertex.Name)

	pterm.Success.Println("codeDAG beginning execution, planned flow printed")
	codeDAG.Visualize()

	if _, err = codeDAG.Run(); err != nil {
		pterm.Error.Printf("Failed to run DAG: %v\n", err)
		return err
	}
	pterm.Success.Println("codeDAG completed in execution successfully")

	return nil
}
