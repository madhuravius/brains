package aws_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockruntimeTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	awsBrains "brains/internal/aws"
	mockBrains "brains/internal/mock"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	b, _ := io.ReadAll(r)
	return string(b)
}

func TestCallAWSBedrockSuccess(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	cfg.SetInvoker(invokerMock)

	expectedBody := []byte(`{"choices":[],"usage":{}}`)
	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{
		Body: expectedBody,
	}, nil)

	req := awsBrains.BedrockRequest{
		Messages: []awsBrains.BedrockMessage{{
			Role: "user",
			Content: []awsBrains.BedrockContent{{
				Type: "text",
				Text: "test",
			}},
		}},
	}
	body, err := cfg.CallAWSBedrock(context.Background(), "model-id", req)
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)
	invokerMock.AssertExpectations(t)
}

func TestCallAWSBedrockError(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	cfg.SetInvoker(invokerMock)

	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(nil, errors.New("invoke error"))

	req := awsBrains.BedrockRequest{}
	_, err := cfg.CallAWSBedrock(context.Background(), "model-id", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoke error")
	invokerMock.AssertExpectations(t)
}

func TestDescribeModelFound(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	inv := &mockBrains.MockInvoker{}
	cfg.SetInvoker(inv)

	modelID := "my-model"
	summary := types.FoundationModelSummary{ModelId: aws.String(modelID)}
	out := &bedrock.ListFoundationModelsOutput{ModelSummaries: []types.FoundationModelSummary{summary}}
	inv.On("ListFoundationModels", mock.Anything, mock.Anything).Return(out, nil)

	got := cfg.DescribeModel(modelID)
	assert.NotNil(t, got)
	assert.Equal(t, modelID, *got.ModelId)
	inv.AssertExpectations(t)
}

func TestDescribeModelNotFound(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	inv := &mockBrains.MockInvoker{}
	cfg.SetInvoker(inv)

	out := &bedrock.ListFoundationModelsOutput{ModelSummaries: []types.FoundationModelSummary{}}
	inv.On("ListFoundationModels", mock.Anything, mock.Anything).Return(out, nil)

	got := cfg.DescribeModel("missing")
	assert.Nil(t, got)
	inv.AssertExpectations(t)
}

func TestCallAWSBedrockConverseSuccess(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	inv := &mockBrains.MockInvoker{}
	cfg.SetInvoker(inv)

	req := awsBrains.BedrockRequest{Messages: []awsBrains.BedrockMessage{{
		Role:    "assistant",
		Content: []awsBrains.BedrockContent{{Type: "text", Text: "hello"}},
	}}}

	text := bedrockruntimeTypes.ContentBlockMemberText{Value: "response"}
	msg := bedrockruntimeTypes.Message{Role: bedrockruntimeTypes.ConversationRoleAssistant, Content: []bedrockruntimeTypes.ContentBlock{&text}}
	outputMember := bedrockruntimeTypes.ConverseOutputMemberMessage{Value: msg}
	convOut := &bedrockruntime.ConverseOutput{Output: &outputMember}
	inv.On("ConverseModel", mock.Anything, mock.Anything).Return(convOut, nil)

	b, err := cfg.CallAWSBedrockConverse(context.Background(), "model-id", req, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("response"), b)
	inv.AssertExpectations(t)
}

func TestCallAWSBedrockConverseError(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	inv := &mockBrains.MockInvoker{}
	cfg.SetInvoker(inv)

	req := awsBrains.BedrockRequest{}
	inv.On("ConverseModel", mock.Anything, mock.Anything).Return(nil, io.ErrUnexpectedEOF)

	_, err := cfg.CallAWSBedrockConverse(context.Background(), "model-id", req, nil)
	assert.Error(t, err)
	inv.AssertExpectations(t)
}

func TestPrintCost(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	assert.NotPanics(t, func() {
		cfg.PrintCost(map[string]any{"prompt_tokens": 10, "completion_tokens": 5}, "modelid")
	})
}

func TestPrintContext(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	assert.NotPanics(t, func() {
		cfg.PrintContext(map[string]any{"prompt_tokens": 300, "completion_tokens": 200})
	})
}

func TestPrintBedrockMessage(t *testing.T) {
	cfg := &awsBrains.AWSConfig{}
	md := "# Header\n\n* item"
	out := captureStdout(func() {
		cfg.PrintBedrockMessage(md)
	})
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Header")
}
