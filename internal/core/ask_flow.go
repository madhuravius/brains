package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/dag"
)

func (a *AskData) SetResearchData(url, data string) {
	if a.ResearchData == nil {
		a.ResearchData = make(map[string]string)
	}
	a.ResearchData[url] = data
}

func (a *AskData) SetRepoMapContext(repoMap string) {
	a.RepoMapContext = repoMap
}

func (a *AskData) SetFileMapData(filePath, fileMapData string) {
	if a.FileMapData == nil {
		a.FileMapData = make(map[string]string)
	}
	a.FileMapData[filePath] = fileMapData
}

func (a *AskData) generateAskFunction(coreConfig *CoreConfig, req *LLMRequest) askDataDAGFunction {
	additionalContext := ""
	for url, data := range a.ResearchData {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	for filePath, fileContents := range a.FileMapData {
		additionalContext += "----- requested file content: " + filePath + "\n\n\n" + fileContents + "\n\n\n" + "------------"
	}
	return func(inputs map[string]string) (string, error) {
		coreConfig.Ask(
			a.RepoMapContext+"\n\nAbove is a mapping of the current repository\n\n"+
				additionalContext+"\n\n\nwere visited above with content if available, you can now return to answering the prompt.\n\n\n"+
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
		ResearchData: make(map[string]string),
	}

	askDAG, err := dag.NewDAG[string, *AskData]("_ask")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		os.Exit(1)
	}

	repoMapVertex := &dag.Vertex[string, *AskData]{
		Name: "repoMap",
		DAG:  askDAG,
		Run:  generateRepoMap(ctx, askData),
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
	promptToSendBedrock := prompt
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock = fmt.Sprintf("%s\n\n%s", logCtx, prompt)
	}
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
	req := aws.BedrockRequest{
		Messages: []aws.BedrockMessage{
			{
				Role: "user",
				Content: []aws.BedrockContent{
					{
						Type: "text",
						Text: promptToSendBedrock,
					},
				},
			},
		},
	}

	respBody, err := c.awsImpl.CallAWSBedrock(ctx, modelID, req)
	if err != nil {
		pterm.Error.Printf("invokeModel error: %v\n", err)
		return false
	}
	var data aws.ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}
	for _, choice := range data.Choices {
		c.logger.LogMessage("[RESPONSE] \n " + choice.Message.Content)
		c.awsImpl.PrintBedrockMessage(choice.Message.Content)
	}
	c.awsImpl.PrintCost(data.Usage, modelID)
	c.awsImpl.PrintContext(data.Usage, modelID)
	return true
}
