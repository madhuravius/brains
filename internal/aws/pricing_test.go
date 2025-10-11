package aws_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/aws"
	mockBrains "github.com/madhuravius/brains/internal/mock"
)

func TestPrintCostOutputsDollar(t *testing.T) {
	cfg := &aws.AWSConfig{}
	cfg.SetPricing([]aws.ModelPricing{
		{
			ModelID:               "amazon.titan-text-lite-v1",
			ModelName:             "Amazon Titan Lite v1",
			InputCostPer1kTokens:  0.0015,
			OutputCostPer1kTokens: 0.003,
		},
	})
	usage := map[string]any{"prompt_tokens": 10, "completion_tokens": 5}
	out := mockBrains.CaptureAllOutput(func() {
		cfg.PrintCost(usage, "amazon.titan-text-lite-v1")
	})
	assert.Contains(t, out, "$")
	assert.Contains(t, out, "amazon.titan-text-lite-v1")
}

func TestPrintContextShowsTokens(t *testing.T) {
	cfg := &aws.AWSConfig{}
	usage := map[string]any{"prompt_tokens": 12., "completion_tokens": 8.}
	out := mockBrains.CaptureAllOutput(func() {
		cfg.PrintContext(usage, "amazon.titan-text-lite-v1")
	})
	assert.Contains(t, out, "20")
	assert.Contains(t, out, "128000")
}

func TestPrintPricingDisplaysModel(t *testing.T) {
	cfg := &aws.AWSConfig{}
	cfg.SetPricing([]aws.ModelPricing{
		{
			ModelID:               "anthropic.claude-v2",
			ModelName:             "Anthropic Claude v2",
			InputCostPer1kTokens:  0.0015,
			OutputCostPer1kTokens: 0.003,
		},
	})
	modelID := "anthropic.claude-v2"
	out := mockBrains.CaptureAllOutput(func() {
		err := cfg.PrintPricing(modelID)
		assert.Nil(t, err)
	})
	assert.Contains(t, out, modelID)
}
