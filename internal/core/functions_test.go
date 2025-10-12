package core_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockruntimeTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/core"
	mockBrains "github.com/madhuravius/brains/internal/mock"
)

func setupCore(t *testing.T) (core.CoreImpl, *mockBrains.MockInvoker) {
	t.Helper()

	awsCfg := &aws.AWSConfig{}
	invoker := &mockBrains.MockInvoker{}
	awsCfg.SetInvoker(invoker)
	awsCfg.SetLogger(&mockBrains.TestLogger{})

	c := core.NewCoreConfig(awsCfg)
	c.SetLogger(&mockBrains.TestLogger{})

	return c, invoker
}

// captureStdout runs a function and returns its combined stdout output.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestValidateBedrockConfiguration_Success(t *testing.T) {
	c, inv := setupCore(t)

	resp := aws.ChatResponse{
		Choices: []aws.ResponseChoice{{
			Message: aws.ResponseMessage{
				Role:    "assistant",
				Content: "All good",
			},
		}},
		Usage: map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 5,
		},
	}
	body, _ := json.Marshal(resp)

	inv.On("InvokeModel", mock.Anything, mock.Anything).
		Return(&bedrockruntime.InvokeModelOutput{Body: body}, nil).
		Once()

	ok := c.ValidateBedrockConfiguration("test-model")
	assert.True(t, ok)
	inv.AssertExpectations(t)
}

func TestValidateBedrockConfiguration_Failure(t *testing.T) {
	c, inv := setupCore(t)

	inv.On("InvokeModel", mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Once()

	ok := c.ValidateBedrockConfiguration("bad-model")
	assert.False(t, ok)
	inv.AssertExpectations(t)
}

func TestAskFlow_Success(t *testing.T) {
	c, inv := setupCore(t)

	inv.
		On("ConverseModel", mock.Anything, mock.Anything).
		Return(&bedrockruntime.ConverseOutput{
			Output: &bedrockruntimeTypes.ConverseOutputMemberMessage{
				Value: bedrockruntimeTypes.Message{
					Role: "assistant",
					Content: []bedrockruntimeTypes.ContentBlock{
						&bedrockruntimeTypes.ContentBlockMemberText{Value: `{
                "name": "assistant", 
                "parameters": {
                  "markdown_summary": "mock",
                  "research_actions": { 
                    "urls_recommended": [], 
                    "files_requested": []
                  } 
                }
              }`},
					},
				},
			},
		}, nil).
		Once()

	inv.
		On("InvokeModel", mock.Anything, mock.Anything).
		Return(&bedrockruntime.InvokeModelOutput{
			Body: []byte(`{
        "choices": [
          {"message": {
            "role": "test",
            "content": "mock response"
          }}
        ],
        "usage": {}
      }`),
		}, nil).
		Once()

	output := captureStdout(func() {
		err := c.AskFlow(context.Background(), &core.LLMRequest{
			Prompt:  "prompt",
			ModelID: "model",
		})
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "mock response")
	inv.AssertExpectations(t)
}

func TestCore_AWSConfig_GetterSetter(t *testing.T) {
	c, _ := setupCore(t)

	newAWS := &aws.AWSConfig{}
	c.SetAWSConfig(newAWS)

	got := c.GetAWSConfig()
	assert.Equal(t, newAWS, got)
}

func TestCore_SetLogger(t *testing.T) {
	c, _ := setupCore(t)
	logger := &mockBrains.TestLogger{}

	// No panic, should assign cleanly
	c.SetLogger(logger)
}
