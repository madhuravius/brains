package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/tools/browser"
	"github.com/madhuravius/brains/internal/tools/repo_map"
)

func generateResearchRun[T Researchable](
	coreConfig *CoreConfig,
	ctx context.Context,
	req *LLMRequest,
	t T,
) askDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		pterm.Info.Println("starting research operation")
		researchActions := coreConfig.Research(req.Prompt, req.ModelID, req.Glob)

		for _, url := range researchActions.UrlsRecommended {
			data, err := browser.FetchWebContext(ctx, url)
			if err != nil {
				pterm.Error.Printf("failed to load url: %v\n", err)
				return "", err
			}
			t.SetResearchData(url, data)
		}
		pterm.Success.Println("research - #1 - loaded browser-based content if applicable")

		for _, fileRequested := range researchActions.FilesRequested {
			data, err := coreConfig.toolsConfig.fsToolConfig.GetFileContents(fileRequested)
			if err != nil {
				pterm.Warning.Printf("failed to load file contents from file requested (%s): %v\n", fileRequested, err)
				continue
			}
			pterm.Info.Printfln("research - #2 - loaded file: %s", fileRequested)
			t.SetFileMapData(fileRequested, data)
		}
		pterm.Success.Println("research - #2 - loaded requested files")

		return "", nil
	}
}

func generateRepoMap[T RepoMappable](
	ctx context.Context,
	t T,
) askDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		repoMap, err := repo_map.BuildRepoMap(ctx, "./")
		if err != nil {
			pterm.Error.Printf("failed to load repo map: %v\n", err)
			return "", err
		}

		t.SetRepoMapContext(repoMap.ToPrompt())

		pterm.Success.Printfln("repoMap successfully constructed: %d files", len(repoMap.Files))
		return "", nil
	}
}

func (c *CoreConfig) enrichWithGlob(glob string) (string, error) {
	addedContext := ""
	if glob != "" {
		var err error
		addedContext, err = c.toolsConfig.fsToolConfig.SetContextFromGlob(glob)
		if err != nil {
			pterm.Error.Printfln("failed to read glob pattern for context: %v", err)
			return "", err
		}
	}
	return addedContext, nil
}

func (c *CoreConfig) Research(prompt, modelID, glob string) *ResearchActions {
	ctx := context.Background()

	promptToSendBedrock := ""
	addedContext, err := c.enrichWithGlob(glob)

	if err != nil {
		return nil
	}
	promptToSendBedrock += addedContext
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock += fmt.Sprintf("%s\n%s\n%s", logCtx, prompt, GeneralResearchActivities)
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

	respBody, err := c.awsImpl.CallAWSBedrockConverse(ctx, modelID, req, researcherToolConfig)
	if err != nil {
		pterm.Error.Printf("converse error: %v\n", err)
		return nil
	}
	c.logger.LogMessage("[RESPONSE FOR RESEARCH] \n " + string(respBody) + "\n\n")

	data, err := ExtractResponse(
		respBody,
		UnwrapFunc[ResearchModelResponse, ResearchModelResponseWithParameters](),
	)
	if err != nil {
		pterm.Error.Printf("unable to ExtractResponse (research): %v\n", err)
		return nil
	}

	return &data.ResearchActions
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

	promptToSendBedrock := ""
	addedContext, err := c.enrichWithGlob(glob)
	if err != nil {
		return nil
	}
	promptToSendBedrock += addedContext
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock += fmt.Sprintf("%s\n%s\n%s", logCtx, prompt, CoderPromptPostProcess)
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

func (c *CoreConfig) ValidateBedrockConfiguration(modelID string) bool {
	ctx := context.Background()
	simpleReq := aws.BedrockRequest{
		Messages: []aws.BedrockMessage{
			{
				Role: "user",
				Content: []aws.BedrockContent{
					{
						Type: "text",
						Text: "This is a health check via API call to make sure a connection to this LLM is established. Please reply with a short three to five word affirmation if you are able to interpret this message that the health check is successful.",
					},
				},
			},
		},
	}
	c.logger.LogMessage("[REQUEST] \n health‑check prompt")
	respBody, err := c.awsImpl.CallAWSBedrock(ctx, modelID, simpleReq)
	if err != nil {
		pterm.Error.Printf("InvokeModel error: %v\n", err)
		return false
	}
	var data aws.ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}
	for _, choice := range data.Choices {
		c.awsImpl.PrintBedrockMessage(choice.Message.Content)
		c.logger.LogMessage("[RESPONSE] \n " + choice.Message.Content)
	}
	c.awsImpl.PrintCost(data.Usage, modelID)
	c.awsImpl.PrintContext(data.Usage, modelID)
	return true
}
