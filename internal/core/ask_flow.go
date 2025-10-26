package core

import (
	"context"
	"fmt"
	"os"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/dag"
)

func (a *AskData) generateAskFunction(coreConfig *CoreConfig, req *LLMRequest) askDataDAGFunction {
	additionalContext := a.generateInitialContextRun()
	return func(inputs map[string]string) (string, error) {
		coreConfig.Ask(
			a.RepoMapContext+"\n\nAbove is a mapping of the current repository\n\n"+
				additionalContext+"\n\n\nis hydrated as initial context, you can now return to answering the prompt.\n\n\n"+
				req.Prompt,
			req.PersonaInstructions,
			req.ModelID,
			req.Glob,
		)

		return "", nil
	}
}

func (c *CoreConfig) AskFlow(ctx context.Context, llmRequest *LLMRequest) error {
	askData := &AskData{
		CommonData: &CommonData{
			ResearchData: make(map[string]string),
		},
	}

	askDAG, err := dag.NewDAG[string, *AskData]("_ask")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		os.Exit(1)
	}

	fileListVertex := &dag.Vertex[string, *AskData]{
		Name: "fileList",
		DAG:  askDAG,
		Run:  generateFileList(c, askData),
	}
	if !c.brainsConfig.GetConfig().ContextConfig.SendFileList {
		fileListVertex.SkipConfig = &dag.SkipVertexConfig{
			Enabled: true,
			Reason:  "send_file_list flag is disabled",
		}
	}
	_ = askDAG.AddVertex(fileListVertex)

	logSummaryVertex := &dag.Vertex[string, *AskData]{
		Name: "logSummary",
		DAG:  askDAG,
		Run:  generateLogSummary(c, llmRequest, ctx, askData),
	}
	if !c.brainsConfig.GetConfig().ContextConfig.SummarizeLogs {
		logSummaryVertex.SkipConfig = &dag.SkipVertexConfig{
			Enabled: true,
			Reason:  "summarize_logs flag is disabled",
		}
	}
	_ = askDAG.AddVertex(logSummaryVertex)

	repoMapVertex := &dag.Vertex[string, *AskData]{
		Name: "repoMap",
		DAG:  askDAG,
		Run:  generateRepoMap(ctx, askData),
	}
	if !c.brainsConfig.GetConfig().ContextConfig.SendAllTags {
		repoMapVertex.SkipConfig = &dag.SkipVertexConfig{
			Enabled: true,
			Reason:  "send_all_tags flag is disabled",
		}
	}
	_ = askDAG.AddVertex(repoMapVertex)

	researchVertex := &dag.Vertex[string, *AskData]{
		Name:        "research",
		DAG:         askDAG,
		Run:         generateResearchRun(c, ctx, llmRequest, askData),
		EnableRetry: true,
	}
	_ = askDAG.AddVertex(researchVertex)

	askVertex := &dag.Vertex[string, *AskData]{
		Name:        "ask",
		DAG:         askDAG,
		Run:         askData.generateAskFunction(c, llmRequest),
		EnableRetry: true,
	}
	_ = askDAG.AddVertex(askVertex)

	askDAG.Connect(fileListVertex.Name, researchVertex.Name)
	askDAG.Connect(logSummaryVertex.Name, researchVertex.Name)
	askDAG.Connect(repoMapVertex.Name, researchVertex.Name)
	askDAG.Connect(researchVertex.Name, askVertex.Name)
	pterm.Success.Println("askDAG beginning execution, planned flow printed")
	askDAG.Visualize()

	if _, err = askDAG.Run(); err != nil {
		pterm.Error.Printf("Failed to run DAG: %v\n", err)
		return err
	}
	pterm.Success.Println("askDAG completed in execution successfully")

	return nil
}

func (c *CoreConfig) Ask(prompt, personaInstructions, modelID, glob string) bool {
	pterm.Info.Println("starting ask operation")
	ctx := context.Background()
	promptToSendBedrock := c.addLogContextToPrompt(prompt)
	if personaInstructions != "" {
		prompt = fmt.Sprintf("%s%s", personaInstructions, prompt)
	}
	c.logger.LogMessage("[REQUEST] \n " + prompt)

	addedContext, err := c.enrichWithGlob(glob)
	if err != nil {
		return false
	}
	if addedContext != "" {
		promptToSendBedrock = fmt.Sprintf("%s%s", prompt, addedContext)
	}
	_, usage, err := c.GenerateBedrockTextResponse(ctx, &LLMRequest{
		Prompt:  promptToSendBedrock,
		ModelID: modelID,
	})
	if err != nil {
		return false
	}

	c.awsImpl.PrintCost(usage, modelID)
	c.awsImpl.PrintContext(usage, modelID)

	return true
}
