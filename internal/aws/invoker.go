package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

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

func (a *AWSConfig) GetInvoker() BedrockInvoker {
	if a.invoker != nil {
		return a.invoker
	}
	return &clientInvoker{
		bedrockruntimeClient: bedrockruntime.NewFromConfig(a.cfg),
		bedrockClient:        bedrock.NewFromConfig(a.cfg),
	}
}
