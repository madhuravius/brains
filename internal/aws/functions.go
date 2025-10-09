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
	pterm.Info.Printfln("size of outbound request: %d", len(body))
	input := &bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	spinner, _ := pterm.DefaultSpinner.Start("loading response from AWS Bedrock")
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

	spinner, _ := pterm.DefaultSpinner.Start("loading response from AWS Bedrock (Converse)")
	resp, err := client.ConverseModel(ctx, input)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	spinner.Success()

	converseOutput, ok := resp.Output.(*bedrockruntimeTypes.ConverseOutputMemberMessage)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from Converse API")
	}

	// attempt json
	for _, block := range converseOutput.Value.Content {
		if toolResultBlock, ok := block.(*bedrockruntimeTypes.ContentBlockMemberToolUse); ok {
			data, err := toolResultBlock.Value.Input.MarshalSmithyDocument()
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	// fall back to text, this may fail
	for _, block := range converseOutput.Value.Content {
		if textBlock, ok := block.(*bedrockruntimeTypes.ContentBlockMemberText); ok {
			pterm.Warning.Println("model returned a text response instead of using the tool. Parsing may be brittle.")
			return []byte(textBlock.Value), nil
		}
	}

	return nil, fmt.Errorf("no tool use or text block found in the response")
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
