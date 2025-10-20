package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/tools/repo_map"
)

func (c *CommonData) SetResearchData(url, data string) {
	if c.ResearchData == nil {
		c.ResearchData = make(map[string]string)
	}
	c.ResearchData[url] = data
}

func (c *CommonData) SetRepoMapContext(repoMap string) {
	c.RepoMapContext = repoMap
}

func (c *CommonData) SetFileListContext(fileListContext string) {
	c.FileListContext = fileListContext
}

func (c *CommonData) SetFileMapData(filePath, fileMapData string) {
	if c.FileMapData == nil {
		c.FileMapData = make(map[string]string)
	}
	c.FileMapData[filePath] = fileMapData
}

func (c *CommonData) SetLogSummaryContext(logSummary string) {
	c.LogSummaryContext = logSummary
}

func (c *CommonData) generateInitialContextRun() string {
	additionalContext := ""
	for url, data := range c.ResearchData {
		additionalContext += "------ scraped content from: " + url + "\n\n\n" + data + "\n\n\n" + "------------"
	}
	for filePath, fileContents := range c.FileMapData {
		additionalContext += "----- requested file content: " + filePath + "\n\n\n" + fileContents + "\n\n\n" + "------------"
	}
	if c.FileListContext != "" {
		additionalContext += "----- requested file list context: \n" + c.FileListContext
	}

	return additionalContext
}

func generateResearchRun[T Researchable](
	coreConfig *CoreConfig,
	ctx context.Context,
	req *LLMRequest,
	t T,
) commonDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		pterm.Info.Println("starting research operation")
		researchActions := coreConfig.Research(req.Prompt, req.ModelID, req.Glob)

		for _, url := range researchActions.UrlsRecommended {
			data, err := coreConfig.toolsConfig.browserToolConfig.FetchWebContext(ctx, url)
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

func generateLogSummary[T LogSummarizable](coreConfig *CoreConfig, llmRequest *LLMRequest, ctx context.Context, t T) commonDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		logCtx := coreConfig.logger.GetLogContext()

		logSummary, usage, err := coreConfig.generateBedrockTextResponse(
			ctx,
			fmt.Sprintf(LogSummary, llmRequest.Prompt, logCtx),
			coreConfig.brainsConfig.GetConfig().Model,
		)
		if err != nil {
			return "", err
		}

		coreConfig.awsImpl.PrintCost(usage, coreConfig.brainsConfig.GetConfig().Model)
		coreConfig.awsImpl.PrintContext(usage, coreConfig.brainsConfig.GetConfig().Model)

		t.SetLogSummaryContext(logSummary)

		pterm.Success.Printfln("log summary successfully constructed")
		return "", nil
	}
}

func generateFileList[T FileListable](
	coreConfig *CoreConfig,
	t T,
) commonDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		fileList, err := coreConfig.toolsConfig.fsToolConfig.GetFileTree("./")
		if err != nil {
			pterm.Error.Printf("failed to load file list: %v\n", err)
			return "", err
		}
		t.SetFileListContext(fileList)

		pterm.Success.Printfln("fileList successfully constructed")
		return "", nil
	}
}

func generateRepoMap[T RepoMappable](
	ctx context.Context,
	t T,
) commonDataDAGFunction {
	return func(inputs map[string]string) (string, error) {
		repoMap, err := repo_map.NewRepoMapConfig(ctx, "./")
		if err != nil {
			pterm.Error.Printf("failed to load repo map: %v\n", err)
			return "", err
		}

		t.SetRepoMapContext(repoMap.ToPrompt())

		pterm.Success.Printfln("repoMap successfully constructed: %d files", repoMap.GetFileCount())
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

func (c *CoreConfig) addLogContextToPrompt(currentPrompt string) string {
	if !c.brainsConfig.GetConfig().ContextConfig.SendLogs {
		return currentPrompt
	}

	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		currentPrompt += fmt.Sprintf("%s\n%s\n%s", logCtx, currentPrompt, GeneralResearchActivities)
	}
	return currentPrompt
}

func (c *CoreConfig) Research(prompt, modelID, glob string) *ResearchActions {
	ctx := context.Background()

	promptToSendBedrock := ""
	addedContext, err := c.enrichWithGlob(glob)

	if err != nil {
		return nil
	}
	promptToSendBedrock = c.addLogContextToPrompt(addedContext)

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

func (c *CoreConfig) ValidateBedrockConfiguration(modelID string) bool {
	ctx := context.Background()
	_, usage, err := c.generateBedrockTextResponse(ctx, HealthCheck, modelID)
	if err != nil {
		return false
	}

	c.awsImpl.PrintCost(usage, modelID)
	c.awsImpl.PrintContext(usage, modelID)

	return true
}

func (c *CoreConfig) generateBedrockTextResponse(ctx context.Context, request, modelID string) (response string, usage map[string]any, err error) {
	simpleReq := aws.BedrockRequest{
		Messages: []aws.BedrockMessage{
			{
				Role: "user",
				Content: []aws.BedrockContent{
					{
						Type: "text",
						Text: request,
					},
				},
			},
		},
	}
	c.logger.LogMessage("[REQUEST] \n health‑check prompt")
	respBody, err := c.awsImpl.CallAWSBedrock(ctx, modelID, simpleReq)
	if err != nil {
		pterm.Error.Printf("invokeModel error: %v\n", err)
		return "", nil, err
	}
	var data aws.ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return "", nil, err
	}

	if len(data.Choices) == 0 {
		pterm.Warning.Printf("no data returned from message response in generateBedrockTextResponse")
		return "", nil, nil
	}

	c.awsImpl.PrintBedrockMessage(data.Choices[0].Message.Content)
	c.logger.LogMessage("[RESPONSE] \n " + data.Choices[0].Message.Content)
	return data.Choices[0].Message.Content, data.Usage, nil
}
