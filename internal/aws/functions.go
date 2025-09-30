package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
)

type BedrockInvoker interface {
	InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error)
	ListFoundationModels(ctx context.Context, input *bedrock.ListFoundationModelsInput) (*bedrock.ListFoundationModelsOutput, error)
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

// ValidateBedrockConfiguration performs a lightweight health‑check against the Bedrock model.
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
	if a.logger != nil {
		a.logger.LogMessage("[REQUEST] health‑check prompt")
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
		if a.logger != nil {
			a.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
		}
	}
	a.printCost(data.Usage)
	a.printContext(data.Usage)
	return true
}

// Ask sends a prompt (optionally prefixed with log context and persona instructions) to Bedrock.
func (a *AWSConfig) Ask(prompt, personaInstructions, addedContext string) bool {
	ctx := context.Background()
	promptToSendBedrock := prompt
	if a.logger != nil {
		if logCtx := a.logger.GetLogContext(); logCtx != "" {
			promptToSendBedrock = fmt.Sprintf("%s\n\n%s", logCtx, prompt)
		}
	}
	if personaInstructions != "" {
		prompt = fmt.Sprintf("%s%s", personaInstructions, prompt)
	}
	if a.logger != nil {
		a.logger.LogMessage("[REQUEST] " + prompt)
	}

	// do not update prompt in place as this will inflate the log
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
		if a.logger != nil {
			a.logger.LogMessage("[RESPONSE] " + choice.Message.Content)
		}
		a.printBedrockMessage(choice.Message.Content)
	}
	a.printCost(data.Usage)
	a.printContext(data.Usage)
	return true
}
