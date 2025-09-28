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

type bedrockSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type bedrockContent struct {
	Type   string         `json:"type"`
	Text   string         `json:"text,omitempty"`
	Source *bedrockSource `json:"source,omitempty"`
}

type bedrockMessage struct {
	Role    string           `json:"role"`
	Content []bedrockContent `json:"content"`
}

type bedrockTool struct {
	Type            string          `json:"type"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	InputSchema     json.RawMessage `json:"input_schema,omitempty"`
	DisplayHeightPx *int            `json:"display_height_px,omitempty"`
	DisplayWidthPx  *int            `json:"display_width_px,omitempty"`
	DisplayNumber   *int            `json:"display_number,omitempty"`
}

type bedrockToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type bedrockRequest struct {
	AnthropicVersion string             `json:"anthropic_version,omitempty"`
	AnthropicBeta    []string           `json:"anthropic_beta,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	System           string             `json:"system,omitempty"`
	Messages         []bedrockMessage   `json:"messages"`
	Temperature      *float64           `json:"temperature,omitempty"`
	TopP             *float64           `json:"top_p,omitempty"`
	TopK             *int               `json:"top_k,omitempty"`
	Tools            []bedrockTool      `json:"tools,omitempty"`
	ToolChoice       *bedrockToolChoice `json:"tool_choice,omitempty"`
	StopSequences    []string           `json:"stop_sequences,omitempty"`
}
