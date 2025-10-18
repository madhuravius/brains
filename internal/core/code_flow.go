package core

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/dag"
)

func (c *CodeData) generateDetermineCodeChangesFunction(coreConfig *CoreConfig, req *LLMRequest) codeDataDAGFunction {
	additionalContext := ""
	for url, data := range c.ResearchData {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	for filePath, fileContents := range c.FileMapData {
		additionalContext += "----- requested file content: " + filePath + "\n\n\n" + fileContents + "\n\n\n" + "------------"
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
		CommonData: &CommonData{
			ResearchData: make(map[string]string),
		},
	}

	codeDAG, err := dag.NewDAG[string, *CodeData]("_code")
	if err != nil {
		pterm.Error.Printf("Failed to initiate DAG: %v\n", err)
		return err
	}

	repoMapVertex := &dag.Vertex[string, *CodeData]{
		Name: "repoMap",
		DAG:  codeDAG,
		Run:  generateRepoMap(ctx, codeData),
	}
	_ = codeDAG.AddVertex(repoMapVertex)

	researchVertex := &dag.Vertex[string, *CodeData]{
		Name: "research",
		DAG:  codeDAG,
		Run:  generateResearchRun(c, ctx, llmRequest, codeData),
	}
	_ = codeDAG.AddVertex(researchVertex)

	determineCodeChangesVertex := &dag.Vertex[string, *CodeData]{
		Name:        "determine_code_changes",
		DAG:         codeDAG,
		Run:         codeData.generateDetermineCodeChangesFunction(c, llmRequest),
		EnableRetry: true,
	}
	_ = codeDAG.AddVertex(determineCodeChangesVertex)

	executeCodeEditsVertex := &dag.Vertex[string, *CodeData]{
		Name:        "execute_code_edits",
		DAG:         codeDAG,
		Run:         codeData.generateExecuteCodeEditsFunction(c),
		EnableRetry: true,
	}
	_ = codeDAG.AddVertex(executeCodeEditsVertex)

	codeDAG.Connect(repoMapVertex.Name, researchVertex.Name)
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

func (c *CoreConfig) ExecuteEditCode(data *CodeModelResponse) bool {
	pterm.Info.Printfln("reviewing each code update, for review one at a time. %d pending updates", len(data.CodeUpdates))

	for updateIdx, update := range data.CodeUpdates {
		pterm.Info.Printfln("updating file: %s (%d/%d)", update.Path, updateIdx+1, len(data.CodeUpdates))

		if _, err := c.toolsConfig.fsToolConfig.UpdateFile(update.Path, update.OldCode, update.NewCode, true); err != nil {
			pterm.Error.Printfln("failed to update %s: %v", update.Path, err)
			return false
		}
	}

	for addIdx, add := range data.AddCodeFiles {
		pterm.Info.Printfln("adding new file: %s (%d/%d)", add.Path, addIdx+1, len(data.AddCodeFiles))
		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Create file %s?", add.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("skipped creation of: %s", add.Path)
			continue
		}
		if err := c.toolsConfig.fsToolConfig.CreateFile(add.Path, add.Content); err != nil {
			pterm.Error.Printfln("failed to write %s: %v", add.Path, err)
			return false
		}
	}

	for remIdx, rem := range data.RemoveCodeFiles {
		pterm.Info.Printfln("removing file: %s (%d/%d)", rem.Path, remIdx+1, len(data.RemoveCodeFiles))

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Delete file %s?", rem.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("skipped deletion of: %s", rem.Path)
			continue
		}
		if err := c.toolsConfig.fsToolConfig.DeleteFile(rem.Path); err != nil {
			pterm.Error.Printfln("failed to write %s: %v", rem.Path, err)
			return false
		}
	}
	return true
}

func (c *CoreConfig) DetermineCodeChanges(prompt, personaInstructions, modelID, glob string) *CodeModelResponse {
	ctx := context.Background()

	promptToSendBedrock, err := c.enrichWithGlob(glob)
	if err != nil {
		return nil
	}
	promptToSendBedrock = c.addLogContextToPrompt(fmt.Sprintf("%s\n%s\n%s", promptToSendBedrock, prompt, CoderPromptPostProcess))

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

	respBody, err := c.awsImpl.CallAWSBedrockConverse(ctx, modelID, req, coderToolConfig)
	if err != nil {
		pterm.Error.Printf("converse error: %v\n", err)
		return nil
	}
	c.logger.LogMessage("[RESPONSE FOR CODE] \n " + string(respBody) + "\n\n")

	data, err := ExtractResponse(
		respBody,
		UnwrapFunc[CodeModelResponse, CodeModelResponseWithParameters](),
	)
	if err != nil {
		pterm.Error.Printf("unable to ExtractResponse (code): %v\n", err)
		return nil
	}

	c.logger.LogMessage("[RESPONSE] \n " + data.MarkdownSummary + "\n\n")
	c.awsImpl.PrintBedrockMessage(data.MarkdownSummary)
	return data

}
