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

func setupCore() (*core.CoreConfig, *mockBrains.MockInvoker) {
	awsCfg := &aws.AWSConfig{}
	invoker := &mockBrains.MockInvoker{}
	awsCfg.SetInvoker(invoker)
	awsCfg.SetLogger(&mockBrains.TestLogger{})
	c := &core.CoreConfig{}
	c.SetAWSConfig(awsCfg)
	c.SetLogger(&mockBrains.TestLogger{})
	return c, invoker
}

func TestCoreFunctions(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*mockBrains.MockInvoker)
		execute       func(*core.CoreConfig) bool
		expectSuccess bool
	}{
		{
			name: "ValidateBedrockConfigurationSuccess",
			setupMock: func(m *mockBrains.MockInvoker) {
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
				m.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{Body: body}, nil)
			},
			execute:       func(c *core.CoreConfig) bool { return c.ValidateBedrockConfiguration("") },
			expectSuccess: true,
		},
		{
			name: "AskSuccess",
			setupMock: func(m *mockBrains.MockInvoker) {
				resp := aws.ChatResponse{
					Choices: []aws.ResponseChoice{{
						Message: aws.ResponseMessage{
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
				body, _ := json.Marshal(resp)
				m.On("InvokeModel", mock.Anything, mock.Anything).Return(&bedrockruntime.InvokeModelOutput{Body: body}, nil)
			},
			execute:       func(c *core.CoreConfig) bool { return c.Ask("prompt", "", "", "") },
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, inv := setupCore()
			tt.setupMock(inv)
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			ok := tt.execute(c)
			_ = w.Close()
			_, _ = io.ReadAll(r)
			os.Stdout = oldStdout
			assert.Equal(t, tt.expectSuccess, ok)
			inv.AssertExpectations(t)
		})
	}
}
