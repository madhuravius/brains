package aws_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brains/internal/aws"
	mockBrains "brains/internal/mock"
)

func TestCallAWSBedrockSuccess(t *testing.T) {
	cfg := &aws.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	cfg.SetInvoker(invokerMock)

	expectedBody := []byte(`{"choices":[],"usage":{}}`)
	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{
		Body: expectedBody,
	}, nil)

	req := aws.BedrockRequest{
		Messages: []aws.BedrockMessage{
			{
				Role: "user",
				Content: []aws.BedrockContent{
					{
						Type: "text",
						Text: "test",
					},
				},
			},
		},
	}
	body, err := cfg.CallAWSBedrock(context.Background(), "model-id", req)
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)
	invokerMock.AssertExpectations(t)
}

func TestCallAWSBedrockError(t *testing.T) {
	cfg := &aws.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	cfg.SetInvoker(invokerMock)

	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(nil, errors.New("invoke error"))

	req := aws.BedrockRequest{}
	_, err := cfg.CallAWSBedrock(context.Background(), "model-id", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoke error")
	invokerMock.AssertExpectations(t)
}

func TestValidateBedrockConfigurationSuccess(t *testing.T) {
	cfg := &aws.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	cfg.SetInvoker(invokerMock)

	cfg.SetLogger(&mockBrains.TestLogger{})

	response := aws.ChatResponse{
		Choices: []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "All good",
				},
			},
		},
		Usage: map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 5,
		},
	}
	respBytes, _ := json.Marshal(response)

	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{
		Body: respBytes,
	}, nil)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ok := cfg.ValidateBedrockConfiguration()
	assert.True(t, ok)

	_ = w.Close()
	_, _ = io.ReadAll(r)
	os.Stdout = oldStdout

	invokerMock.AssertExpectations(t)
}
