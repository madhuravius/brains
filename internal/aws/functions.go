package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockruntimeTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
)

const (
	inputPricePerK  = 0.00015
	outputPricePerK = 0.0006
	tokenLimit      = 128000
)

func (a *AWSConfig) DescribeModel(model string) *types.FoundationModelSummary {
	client := a.GetInvoker()
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
	client := a.GetInvoker()
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Bedrock request: %w", err)
	}
	pterm.Info.Printfln("Size of outbound request: %d", len(body))
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

func (a *AWSConfig) CallAWSBedrockConverse(
	ctx context.Context,
	modelID string,
	req BedrockRequest,
	toolConfig *bedrockruntimeTypes.ToolConfiguration,
) ([]byte, error) {
	client := a.GetInvoker()

	messages := []bedrockruntimeTypes.Message{}
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

	if toolConfig != nil {
		input.ToolConfig = toolConfig
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
	var text *bedrockruntimeTypes.ContentBlockMemberText
	for _, responseBlock := range responseText.Value.Content {
		returnedText, okText := responseBlock.(*bedrockruntimeTypes.ContentBlockMemberText)
		if !okText {
			continue
		} else {
			text = returnedText
			break
		}
	}
	if text == nil {
		return nil, fmt.Errorf("empty response content (no text)")
	}
	return []byte(text.Value), nil
}

func (a *AWSConfig) PrintCost(usage map[string]any) {
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
	cost := (float64(promptTokens) / 1000.0 * inputPricePerK) + (float64(completionTokens) / 1000.0 * outputPricePerK)
	pterm.Info.Printf("Estimated cost for this request: $%.6f (prompt tokens: %d, completion tokens: %d)\n", cost, promptTokens, completionTokens)
}

func (a *AWSConfig) PrintContext(usage map[string]any) {
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
	pterm.Info.Printf("Current context used: %d tokens (limit %d)\n", total, tokenLimit)
}

func (a *AWSConfig) PrintBedrockMessage(content string) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
		glamour.WithColorProfile(termenv.ANSI256),
	)
	result, _ := r.Render(content)
	fmt.Println(result)
}
