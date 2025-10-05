package aws

import "encoding/json"

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage map[string]any
}

type BedrockSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type BedrockContent struct {
	Type   string         `json:"type"`
	Text   string         `json:"text,omitempty"`
	Source *BedrockSource `json:"source,omitempty"`
}

type BedrockMessage struct {
	Role    string           `json:"role"`
	Content []BedrockContent `json:"content"`
}

type BedrockTool struct {
	Type            string          `json:"type"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	InputSchema     json.RawMessage `json:"input_schema,omitempty"`
	DisplayHeightPx *int            `json:"display_height_px,omitempty"`
	DisplayWidthPx  *int            `json:"display_width_px,omitempty"`
	DisplayNumber   *int            `json:"display_number,omitempty"`
}

type BedrockToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type BedrockRequest struct {
	AnthropicVersion string             `json:"anthropic_version,omitempty"`
	AnthropicBeta    []string           `json:"anthropic_beta,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	System           string             `json:"system,omitempty"`
	Messages         []BedrockMessage   `json:"messages"`
	Temperature      *float64           `json:"temperature,omitempty"`
	TopP             *float64           `json:"top_p,omitempty"`
	TopK             *int               `json:"top_k,omitempty"`
	Tools            []BedrockTool      `json:"tools,omitempty"`
	ToolChoice       *BedrockToolChoice `json:"tool_choice,omitempty"`
	StopSequences    []string           `json:"stop_sequences,omitempty"`
}

type CodeUpdate struct {
	Path    string `json:"path"`
	OldCode string `json:"old_code"`
	NewCode string `json:"new_code"`
}

type AddCodeFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type RemoveCodeFile struct {
	Path string `json:"path"`
}

type CodeModelResponse struct {
	CodeUpdates     []CodeUpdate     `json:"code_updates"`
	AddCodeFiles    []AddCodeFile    `json:"add_code_files,omitempty"`
	RemoveCodeFiles []RemoveCodeFile `json:"remove_code_files,omitempty"`
}
