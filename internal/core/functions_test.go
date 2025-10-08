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

func TestExtractCodeModelResponse_ArrayOfCodeModelResponse(t *testing.T) {
	resp := []core.CodeModelResponse{{MarkdownSummary: "array simple"}}
	data, _ := json.Marshal(resp)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.NoError(t, err)
	assert.Equal(t, "array simple", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_ArrayOfCodeModelResponseWithParameters(t *testing.T) {
	wrapper := []core.CodeModelResponseWithParameters{{
		Name:       "test",
		Parameters: core.CodeModelResponse{MarkdownSummary: "array with params"},
	}}
	data, _ := json.Marshal(wrapper)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.NoError(t, err)
	assert.Equal(t, "array with params", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_CodeModelResponse(t *testing.T) {
	resp := core.CodeModelResponse{MarkdownSummary: "object simple"}
	data, _ := json.Marshal(resp)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.NoError(t, err)
	assert.Equal(t, "object simple", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_CodeModelResponseWithParameters(t *testing.T) {
	wrapper := core.CodeModelResponseWithParameters{
		Name:       "test",
		Parameters: core.CodeModelResponse{MarkdownSummary: "object with params"},
	}
	data, _ := json.Marshal(wrapper)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.NoError(t, err)
	assert.Equal(t, "object with params", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_Unrecognized(t *testing.T) {
	data := []byte(`{"unexpected":"value"}`)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.Nil(t, cmr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized or empty JSON structure")
}

func TestExtractCodeModelResponse_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json`)
	cmr, err := core.ExtractCodeModelResponse(data)
	assert.Nil(t, cmr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}
