package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockInvoker interface {
	InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error)
}
