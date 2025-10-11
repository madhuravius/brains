package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	bedrockruntimeTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	brainsConfig "github.com/madhuravius/brains/internal/config"
)

type AWSImpl interface {
	CallAWSBedrock(ctx context.Context, modelID string, req BedrockRequest) ([]byte, error)
	CallAWSBedrockConverse(
		ctx context.Context,
		modelID string,
		req BedrockRequest,
		toolConfig *bedrockruntimeTypes.ToolConfiguration,
	) ([]byte, error)
	DescribeModel(model string) *types.FoundationModelSummary
	GetConfig() aws.Config
	PrintBedrockMessage(content string)
	PrintContext(usage map[string]any, modelID string)
	PrintCost(usage map[string]any, modelID string)
	PrintPricing(modelID string) error
	SetAndValidateCredentials() bool
	SetLogger(l brainsConfig.SimpleLogger)
	SetPricing(pricing []ModelPricing)
}

type AggregatedModelPricing struct {
	ModelName             string
	ModelID               string
	ProviderName          string
	InputCostPer1kTokens  float64
	OutputCostPer1kTokens float64
}

type ModelPricing struct {
	ModelID               string  `json:"ModelID"`
	ModelName             string  `json:"ModelName"`
	InputCostPer1kTokens  float64 `json:"InputCostPer1kTokens"`
	OutputCostPer1kTokens float64 `json:"OutputCostPer1kTokens"`
}

type AWSConfig struct {
	cfg     aws.Config
	region  string
	invoker BedrockInvoker
	logger  brainsConfig.SimpleLogger

	pricing []ModelPricing
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseChoice struct {
	Message ResponseMessage `json:"message"`
}

type ChatResponse struct {
	Choices []ResponseChoice `json:"choices"`
	Usage   map[string]any
}

type BedrockSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type BedrockContent struct {
	Type   string         `json:"type"`
	Text   string         `json:"text,omitempty"`
	Source *BedrockSource `json:"source,omitempty"`
}

type BedrockMessage struct {
	Role    string           `json:"role"`
	Content []BedrockContent `json:"content"`
}

type BedrockTool struct {
	Type            string          `json:"type"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	InputSchema     json.RawMessage `json:"input_schema,omitempty"`
	DisplayHeightPx *int            `json:"display_height_px,omitempty"`
	DisplayWidthPx  *int            `json:"display_width_px,omitempty"`
	DisplayNumber   *int            `json:"display_number,omitempty"`
}

type BedrockToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type BedrockRequest struct {
	AnthropicVersion string             `json:"anthropic_version,omitempty"`
	AnthropicBeta    []string           `json:"anthropic_beta,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	System           string             `json:"system,omitempty"`
	Messages         []BedrockMessage   `json:"messages"`
	Temperature      *float64           `json:"temperature,omitempty"`
	TopP             *float64           `json:"top_p,omitempty"`
	TopK             *int               `json:"top_k,omitempty"`
	Tools            []BedrockTool      `json:"tools,omitempty"`
	ToolChoice       *BedrockToolChoice `json:"tool_choice,omitempty"`
	StopSequences    []string           `json:"stop_sequences,omitempty"`
}

type CodeUpdate struct {
	Path    string `json:"path"`
	OldCode string `json:"old_code"`
	NewCode string `json:"new_code"`
}

type AddCodeFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type RemoveCodeFile struct {
	Path string `json:"path"`
}
