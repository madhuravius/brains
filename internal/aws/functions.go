package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
)

func (a *AWSConfig) CallAWSBedrock(ctx context.Context, modelID string, req bedrockRequest) ([]byte, error) {
	client := bedrockruntime.NewFromConfig(a.cfg)
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
	resp, err := client.InvokeModel(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
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

func (a *AWSConfig) printBedrockMessage(content string) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	result, _ := r.Render(content)
	fmt.Println(result)
}

func (a *AWSConfig) ValidateBedrockConfiguration() bool {
	ctx := context.Background()
	simpleReq := bedrockRequest{
		Messages: []bedrockMessage{
			{
				Role: "user",
				Content: []bedrockContent{
					{
						Type: "text",
						Text: "This is a health check via API call to make sure a connection to this LLM is established. Please reply with a short three to five word affirmation if you are able to interpret this message that the health check is successful.",
					},
				},
			},
		},
	}
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
	}
	a.printCost(data.Usage)
	return true
}

func (a *AWSConfig) Ask(prompt string) bool {
	ctx := context.Background()
	req := bedrockRequest{
		Messages: []bedrockMessage{
			{
				Role: "user",
				Content: []bedrockContent{
					{
						Type: "text",
						Text: prompt,
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
		a.printBedrockMessage(choice.Message.Content)
	}
	a.printCost(data.Usage)
	return true
}
