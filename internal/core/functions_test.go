package core_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"brains/internal/aws"
	"brains/internal/core"
	mockBrains "brains/internal/mock"
)

func TestValidateBedrockConfigurationSuccess(t *testing.T) {
	awsCfg := &aws.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	awsCfg.SetInvoker(invokerMock)
	awsCfg.SetLogger(&mockBrains.TestLogger{})

	cfg := &core.CoreConfig{}
	cfg.SetAWSConfig(awsCfg)
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

	ok := cfg.ValidateBedrockConfiguration("")
	assert.True(t, ok)

	_ = w.Close()
	_, _ = io.ReadAll(r)
	os.Stdout = oldStdout

	invokerMock.AssertExpectations(t)
}

func TestAskSuccess(t *testing.T) {
	awsCfg := &aws.AWSConfig{}
	invokerMock := &mockBrains.MockInvoker{}
	awsCfg.SetInvoker(invokerMock)
	awsCfg.SetLogger(&mockBrains.TestLogger{})

	cfg := &core.CoreConfig{}
	cfg.SetAWSConfig(awsCfg)
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
					Content: "Reply",
				},
			},
		},
		Usage: map[string]any{
			"prompt_tokens":     1,
			"completion_tokens": 1,
		},
	}
	respBytes, _ := json.Marshal(response)

	invokerMock.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{
		Body: respBytes,
	}, nil)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ok := cfg.Ask("prompt", "", "", "")
	assert.True(t, ok)

	_ = w.Close()
	_, _ = io.ReadAll(r)
	os.Stdout = oldStdout

	invokerMock.AssertExpectations(t)
}
