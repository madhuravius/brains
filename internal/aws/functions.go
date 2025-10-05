package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockruntimeTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const CoderPromptPostProcess = `
You are a code-editing assistant.

Evaluate the most recent recommendations in the context provided. These should be implemented in code now.

Return **only JSON** describing the changes, no explanations.

Return JSON in this format:
{
  "code_updates": [
    {"path": "example_file_name.go", "old_code": "...", "new_code": "..."}
  ],
    "add_code_files": [
    {"path": "new_file.go", "content": "..."}
  ],
    "remove_code_files": [
    {"path": "old_file.go"}
  ]
}
`

type BedrockInvoker interface {
	InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error)
	ListFoundationModels(ctx context.Context, input *bedrock.ListFoundationModelsInput) (*bedrock.ListFoundationModelsOutput, error)
	ConverseModel(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error)
}

type clientInvoker struct {
	bedrockruntimeClient *bedrockruntime.Client
	bedrockClient        *bedrock.Client
}

func (c *clientInvoker) InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error) {
	return c.bedrockruntimeClient.InvokeModel(ctx, input)
}

func (c *clientInvoker) ListFoundationModels(ctx context.Context, input *bedrock.ListFoundationModelsInput) (*bedrock.ListFoundationModelsOutput, error) {
	return c.bedrockClient.ListFoundationModels(ctx, input)
}

func (c *clientInvoker) ConverseModel(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	return c.bedrockruntimeClient.Converse(ctx, input)
}

func (a *AWSConfig) SetInvoker(invoker BedrockInvoker) {
	a.invoker = invoker
}

func (a *AWSConfig) getInvoker() BedrockInvoker {
	if a.invoker != nil {
		return a.invoker
	}
	return &clientInvoker{
		bedrockruntimeClient: bedrockruntime.NewFromConfig(a.cfg),
		bedrockClient:        bedrock.NewFromConfig(a.cfg),
	}
}

func (a *AWSConfig) DescribeModel(model string) *types.FoundationModelSummary {
	client := a.getInvoker()
	out, err := client.ListFoundationModels(context.Background(),
		&bedrock.ListFoundationModelsInput{})
	if err != nil {
		pterm.Error.Printf("list models: %v\n", err)
		return nil
	}
	for _, m := range out.ModelSummaries {
		if *m.ModelId == model {
			return &m
		}
	}
	pterm.Error.Printf("model %s not found\n", model)
	return nil
}

func (a *AWSConfig) CallAWSBedrock(ctx context.Context, modelID string, req BedrockRequest) ([]byte, error) {
	client := a.getInvoker()
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Bedrock request: %w", err)
	}
	input := &bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	spinner, _ := pterm.DefaultSpinner.Start("Loading response from AWS Bedrock")
	resp, err := client.InvokeModel(ctx, input)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	spinner.Success()
	return resp.Body, nil
}

func (a *AWSConfig) CallAWSBedrockConverse(ctx context.Context, modelID string, req BedrockRequest) ([]byte, error) {
	client := a.getInvoker()

	var messages []bedrockruntimeTypes.Message
	for _, m := range req.Messages {
		var content bedrockruntimeTypes.ContentBlockMemberText
		if len(m.Content) > 0 && m.Content[0].Type == "text" {
			content = bedrockruntimeTypes.ContentBlockMemberText{
				Value: m.Content[0].Text,
			}
		}
		message := bedrockruntimeTypes.Message{
			Content: []bedrockruntimeTypes.ContentBlock{&content},
			Role:    bedrockruntimeTypes.ConversationRole(m.Role),
		}
		messages = append(messages, message)
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  aws.String(modelID),
		Messages: messages,
	}

	spinner, _ := pterm.DefaultSpinner.Start("Loading response from AWS Bedrock (Converse)")
	resp, err := client.ConverseModel(ctx, input)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	spinner.Success()

	responseText, ok := resp.Output.(*bedrockruntimeTypes.ConverseOutputMemberMessage)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}
	if len(responseText.Value.Content) == 0 {
		return nil, fmt.Errorf("empty response content")
	}
	text, ok := responseText.Value.Content[0].(*bedrockruntimeTypes.ContentBlockMemberText)
	if !ok {
		return nil, fmt.Errorf("unexpected content type")
	}
	return []byte(text.Value), nil
}

func (a *AWSConfig) printCost(usage map[string]any) {
	promptTokens, completionTokens := 0, 0
	if v, ok := usage["prompt_tokens"]; ok {
		if val, okFloat := v.(float64); okFloat {
			promptTokens = int(val)
		}
	}
	if v, ok := usage["completion_tokens"]; ok {
		if val, okFloat := v.(float64); okFloat {
			completionTokens = int(val)
		}
	}
	inputPricePerK := 0.00015
	outputPricePerK := 0.0006
	cost := (float64(promptTokens) / 1000.0 * inputPricePerK) + (float64(completionTokens) / 1000.0 * outputPricePerK)
	pterm.Info.Printf("Estimated cost for this request: $%.6f (prompt tokens: %d, completion tokens: %d)\n", cost, promptTokens, completionTokens)
}

func (a *AWSConfig) printContext(usage map[string]any) {
	promptTokens, completionTokens := 0, 0
	if v, ok := usage["prompt_tokens"]; ok {
		if val, okFloat := v.(float64); okFloat {
			promptTokens = int(val)
		}
	}
	if v, ok := usage["completion_tokens"]; ok {
		if val, okFloat := v.(float64); okFloat {
			completionTokens = int(val)
		}
	}
	total := promptTokens + completionTokens
	const tokenLimit = 128000
	pterm.Info.Printf("Current context used: %d tokens (limit %d)\n", total, tokenLimit)
}

func (a *AWSConfig) printBedrockMessage(content string) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	result, _ := r.Render(content)
	fmt.Println(result)
}

func (a *AWSConfig) ValidateBedrockConfiguration() bool {
	ctx := context.Background()
	simpleReq := BedrockRequest{
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContent{
					{
						Type: "text",
						Text: "This is a health check via API call to make sure a connection to this LLM is established. Please reply with a short three to five word affirmation if you are able to interpret this message that the health check is successful.",
					},
				},
			},
		},
	}
	a.logger.LogMessage("[REQUEST] health‑check prompt")
	respBody, err := a.CallAWSBedrock(ctx, a.defaultBedrockModel, simpleReq)
	if err != nil {
		pterm.Error.Printf("InvokeModel error: %v\n", err)
		return false
	}
	var data ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}
	for _, choice := range data.Choices {
		a.printBedrockMessage(choice.Message.Content)
		a.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
	}
	a.printCost(data.Usage)
	a.printContext(data.Usage)
	return true
}

func (a *AWSConfig) Ask(prompt, personaInstructions, addedContext string) bool {
	ctx := context.Background()
	promptToSendBedrock := prompt
	if logCtx := a.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock = fmt.Sprintf("%s\n\n%s", logCtx, prompt)
	}
	if personaInstructions != "" {
		prompt = fmt.Sprintf("%s%s", personaInstructions, prompt)
	}
	a.logger.LogMessage("[REQUEST] " + prompt)

	if addedContext != "" {
		promptToSendBedrock = fmt.Sprintf("%s%s", prompt, addedContext)
	}
	req := BedrockRequest{
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContent{
					{
						Type: "text",
						Text: promptToSendBedrock,
					},
				},
			},
		},
	}

	respBody, err := a.CallAWSBedrock(ctx, a.defaultBedrockModel, req)
	if err != nil {
		pterm.Error.Printf("InvokeModel error: %v\n", err)
		return false
	}
	var data ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}
	for _, choice := range data.Choices {
		a.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
		a.printBedrockMessage(choice.Message.Content)
	}
	a.printCost(data.Usage)
	a.printContext(data.Usage)
	return true
}

func (a *AWSConfig) Code(prompt, personaInstructions, addedContext string) bool {
	if !a.Ask(prompt, personaInstructions, addedContext) {
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
	if logCtx := a.logger.GetLogContext(); logCtx != "" {
		promptToSendBedrock += fmt.Sprintf("%s\n\n%s", logCtx, CoderPromptPostProcess)
	}

	req := BedrockRequest{
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContent{
					{
						Type: "text",
						Text: promptToSendBedrock,
					},
				},
			},
		},
	}

	respBody, err := a.CallAWSBedrockConverse(ctx, a.defaultBedrockModel, req)
	if err != nil {
		pterm.Error.Printf("Converse error: %v\n", err)
		return false
	}
	var data ChatResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing Bedrock Body): %v\n", err)
		return false
	}
	if len(data.Choices) == 0 {
		pterm.Error.Printf("On JSON Response for code updates did not get sufficient response")
		return false
	}

	a.logger.LogMessage("[RESPONSE FOR CODE] " + data.Choices[0].Message.Content)

	var updates CodeModelResponse
	if err := json.Unmarshal([]byte(data.Choices[0].Message.Content), &updates); err != nil {
		pterm.Error.Printf("Json Unmarshal error (when parsing codeModelUpdates Body): %v\n", err)
		return false
	}

	for updateIdx, update := range updates.CodeUpdates {
		pterm.Info.Printfln("Updating file: %s (%d/%d)", update.Path, updateIdx+1, len(updates.CodeUpdates))

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

	for addIdx, add := range updates.AddCodeFiles {
		pterm.Info.Printfln("Adding new file: %s (%d/%d)", add.Path, addIdx+1, len(updates.AddCodeFiles))

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

	for remIdx, rem := range updates.RemoveCodeFiles {
		pterm.Info.Printfln("Removing file: %s (%d/%d)", rem.Path, remIdx+1, len(updates.RemoveCodeFiles))

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

	a.printCost(data.Usage)
	a.printContext(data.Usage)

	return true
}
