package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"

	"brains/internal/aws"
)

func (c *CoreConfig) Ask(prompt, personaInstructions, addedContext, modelID string) bool {
	ctx := context.Background()
	promptToSendBedrock := prompt
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock = fmt.Sprintf("%s\n\n%s", logCtx, prompt)
	}
	if personaInstructions != "" {
		prompt = fmt.Sprintf("%s%s", personaInstructions, prompt)
	}
	c.logger.LogMessage("[REQUEST] " + prompt)

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

func (c *CoreConfig) Code(prompt, personaInstructions, modelID, addedContext string) bool {
	if !c.Ask(prompt, personaInstructions, modelID, addedContext) {
		return false
	}
	pterm.Info.Printf("will edit files in place for those listed above\n")
	result, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Continue with file edits?").Show()
	if !result {
		pterm.Warning.Printf("refused to continue, breaking early")
		return false
	}

	promptToSendBedrock := ""

	ctx := context.Background()
	promptToSendBedrock += addedContext
	if logCtx := c.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock += fmt.Sprintf("%s\n\n%s", logCtx, CoderPromptPostProcess)
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

	respBody, err := c.awsConfig.CallAWSBedrockConverse(ctx, modelID, req)
	if err != nil {
		pterm.Error.Printf("Converse error: %v\n", err)
		return false
	}
	c.logger.LogMessage("[RESPONSE FOR CODE] " + string(respBody))

	var data aws.CodeModelResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}

	for updateIdx, update := range data.CodeUpdates {
		pterm.Info.Printfln("Updating file: %s (%d/%d)", update.Path, updateIdx+1, len(data.CodeUpdates))

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(update.OldCode, update.NewCode, false)
		diffText := dmp.DiffPrettyText(diffs)

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
		fmt.Println(renderedDiff)

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Apply changes to %s?", update.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped: %s", update.Path)
			continue
		}

		if err := os.WriteFile(update.Path, []byte(update.NewCode), 0644); err != nil {
			pterm.Error.Printfln("Failed to write %s: %v", update.Path, err)
			continue
		}
		pterm.Success.Printfln("Updated %s successfully", update.Path)
	}

	for addIdx, add := range data.AddCodeFiles {
		pterm.Info.Printfln("Adding new file: %s (%d/%d)", add.Path, addIdx+1, len(data.AddCodeFiles))

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Create file %s?", add.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped creation of: %s", add.Path)
			continue
		}

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain("", add.Content, false)
		diffText := dmp.DiffPrettyText(diffs)

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
		fmt.Println(renderedDiff)

		if err := os.WriteFile(add.Path, []byte(add.Content), 0644); err != nil {
			pterm.Error.Printfln("Failed to create %s: %v", add.Path, err)
			continue
		}
		pterm.Success.Printfln("Created %s successfully", add.Path)
	}

	for remIdx, rem := range data.RemoveCodeFiles {
		pterm.Info.Printfln("Removing file: %s (%d/%d)", rem.Path, remIdx+1, len(data.RemoveCodeFiles))

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Delete file %s?", rem.Path)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped deletion of: %s", rem.Path)
			continue
		}

		oldContentBytes, readErr := os.ReadFile(rem.Path)
		if readErr != nil {
			pterm.Error.Printfln("Failed to read %s for diff: %v", rem.Path, readErr)
		} else {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(oldContentBytes), "", false)
			diffText := dmp.DiffPrettyText(diffs)

			r, _ := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(100),
			)
			renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
			fmt.Println(renderedDiff)
		}

		if err := os.Remove(rem.Path); err != nil {
			pterm.Error.Printfln("Failed to delete %s: %v", rem.Path, err)
			continue
		}
		pterm.Success.Printfln("Deleted %s successfully", rem.Path)
	}

	return true
}
