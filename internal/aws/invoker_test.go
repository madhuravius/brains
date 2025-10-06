package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockBrains "brains/internal/mock"
)

func TestClientInvokerExecutesAllMethods(t *testing.T) {
	cfg := &AWSConfig{}
	mockInv := &mockBrains.MockInvoker{}
	cfg.SetInvoker(mockInv)

	mockInv.
		On("InvokeModel", mock.Anything, mock.Anything).
		Return((*bedrockruntime.InvokeModelOutput)(nil), errors.New("invoke error"))
	mockInv.
		On("ListFoundationModels", mock.Anything, mock.Anything).
		Return((*bedrock.ListFoundationModelsOutput)(nil), errors.New("list error"))
	mockInv.
		On("ConverseModel", mock.Anything, mock.Anything).
		Return((*bedrockruntime.ConverseOutput)(nil), errors.New("converse error"))

	inv := cfg.GetInvoker()
	assert.NotNil(t, inv)

	_, err := inv.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{})
	assert.Error(t, err)

	_, err = inv.ListFoundationModels(context.Background(), &bedrock.ListFoundationModelsInput{})
	assert.Error(t, err)

	_, err = inv.ConverseModel(context.Background(), &bedrockruntime.ConverseInput{})
	assert.Error(t, err)

	mockInv.AssertExpectations(t)
}
