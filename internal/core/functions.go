package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"

	"brains/internal/aws"
)

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

func (c *CoreConfig) Ask(prompt, personaInstructions, modelID, glob string) bool {
	ctx := context.Background()
	promptToSendBedrock := prompt
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock = fmt.Sprintf("%s\n\n%s", logCtx, prompt)
	}
	if personaInstructions != "" {
		prompt = fmt.Sprintf("%s%s", personaInstructions, prompt)
	}
	c.logger.LogMessage("[REQUEST] " + prompt)

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

	respBody, err := c.awsConfig.CallAWSBedrock(ctx, modelID, req)
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
		c.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
		c.awsConfig.PrintBedrockMessage(choice.Message.Content)
	}
	c.awsConfig.PrintCost(data.Usage)
	c.awsConfig.PrintContext(data.Usage)
	return true
}

func (c *CoreConfig) Code(prompt, personaInstructions, modelID, glob string) bool {
	ctx := context.Background()

	promptToSendBedrock := ""
	addedContext, err := c.enrichWithGlob(glob)
	if err != nil {
		return false
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

	respBody, err := c.awsConfig.CallAWSBedrockConverse(ctx, modelID, req, coderToolConfig)
	if err != nil {
		pterm.Error.Printf("Converse error: %v\n", err)
		return false
	}
	c.logger.LogMessage("[RESPONSE FOR CODE] " + string(respBody) + "\n\n")

	// below goofy-ass logic is required as the llm will not consistently reply back in an array or single object
	var raw json.RawMessage
	if err := json.Unmarshal(respBody, &raw); err != nil {
		pterm.Error.Printf("Json Unmarshal error: %v\n", err)
		return false
	}
	var data CodeModelResponse
	if raw[0] == '[' { // it's an array
		var arr []CodeModelResponse
		if err := json.Unmarshal(raw, &arr); err != nil {
			pterm.Error.Printf("Array unmarshal error: %v\n", err)
			return false
		}
		data = arr[0]
	} else { // single object
		if err := json.Unmarshal(raw, &data); err != nil {
			pterm.Error.Printf("Object unmarshal error: %v\n", err)
			return false
		}
	}

	c.logger.LogMessage("[RESPONSE] " + data.MarkdownSummary + "\n\n")
	c.awsConfig.PrintBedrockMessage(data.MarkdownSummary)

	pterm.Info.Printfln("Reviewing each code update, for review one at a time. %d pending updates", len(data.CodeUpdates))

	for updateIdx, update := range data.CodeUpdates {
		pterm.Info.Printfln("Updating file: %s (%d/%d)", update.Path, updateIdx+1, len(data.CodeUpdates))

		if _, err := c.toolsConfig.fsToolConfig.UpdateFile(update.Path, update.OldCode, update.NewCode, true); err != nil {
			pterm.Error.Printfln("Failed to update %s: %v", update.Path, err)
			return false
		}
	}

	for addIdx, add := range data.AddCodeFiles {
		pterm.Info.Printfln("Adding new file: %s (%d/%d)", add.Path, addIdx+1, len(data.AddCodeFiles))
		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Create file %s?", add.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped creation of: %s", add.Path)
			continue
		}
		if err := c.toolsConfig.fsToolConfig.CreateFile(add.Path, add.Content); err != nil {
			pterm.Error.Printfln("Failed to write %s: %v", add.Path, err)
			return false
		}
	}

	for remIdx, rem := range data.RemoveCodeFiles {
		pterm.Info.Printfln("Removing file: %s (%d/%d)", rem.Path, remIdx+1, len(data.RemoveCodeFiles))

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Delete file %s?", rem.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped deletion of: %s", rem.Path)
			continue
		}
		if err := c.toolsConfig.fsToolConfig.DeleteFile(rem.Path); err != nil {
			pterm.Error.Printfln("Failed to write %s: %v", rem.Path, err)
			return false
		}
	}

	return true
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
	c.logger.LogMessage("[REQUEST] health‑check prompt")
	respBody, err := c.awsConfig.CallAWSBedrock(ctx, modelID, simpleReq)
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
		c.awsConfig.PrintBedrockMessage(choice.Message.Content)
		c.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
	}
	c.awsConfig.PrintCost(data.Usage)
	c.awsConfig.PrintContext(data.Usage)
	return true
}
