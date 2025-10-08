package core

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const GeneralResearchActivities = `
You are a research augmentation assistant.

Your sole purpose is to support the parent prompt by gathering *only the minimal additional information necessary* to complete its specific task accurately.

You must:
- Use existing URLs provided in the prompt only if they are relevant and necessary.
- Strictly limit all research and retrieval to the direct requirements of the parent task.
- Avoid redundant or speculative research outside the task scope.

Return **only JSON**, with the following exact schema:
{
  "markdown_summary": "string",
  "research_actions": {
    "urls_recommended": ["string"],
  }
}

- "markdown_summary": a concise markdown summary of what was found.
- "research_actions": explicitly list which URLs need to be searched

Do NOT include explanations, reasoning steps, or extra text outside of the JSON.
Return valid JSON only.
`

var researcherToolConfig = &types.ToolConfiguration{
	Tools: []types.Tool{
		&types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				Name:        aws.String("researcher"),
				Description: aws.String("Gather minimal external information to support the parent prompt’s specific task"),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: document.NewLazyDocument(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"markdown_summary": map[string]any{
								"type":        "string",
								"description": "Concise markdown summary of the relevant information found.",
							},
							"research_actions": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"urls_recommended": map[string]any{
										"type":        "array",
										"items":       map[string]any{"type": "string"},
										"description": "List of URLs that should be researched to supplement the parent prompt.",
									},
								},
								"required": []string{"urls_recommended"},
							},
						},
						"required": []string{"markdown_summary", "research_actions"},
					}),
				},
			},
		},
	},
	ToolChoice: &types.ToolChoiceMemberAny{
		Value: types.AnyToolChoice{},
	},
}

const CoderPromptPostProcess = `
You are a code-editing assistant.

Ensure that when outputting the file with changes, you absolutely include the full file output instead of snipping sections and adding comments such as "// (rest of the code remains unchanged)" as that can cause the code to be overwritten instead of skipped.

Return **only JSON** describing the changes, no explanations. 

Return JSON in this format, strictly with the following schema:
{
  "markdown_summary": "string",
  "code_updates": [
    {"path": "string", "old_code": "string", "new_code": "string"}
  ],
  "add_code_files": [
    {"path": "string", "content": "string"}
  ],
  "remove_code_files": [
    {"path": "string"}
  ]
}

Do NOT include extra text or commentary. Only return JSON.
Analyze the code changes and generate the JSON accordingly.
`

var coderToolConfig = &types.ToolConfiguration{
	Tools: []types.Tool{
		&types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				Name:        aws.String("coder"),
				Description: aws.String("Generate code changes in a specific schema"),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: document.NewLazyDocument(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"markdown_summary": map[string]any{
								"type": "string",
							},
							"code_updates": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"path":     map[string]any{"type": "string"},
										"old_code": map[string]any{"type": "string"},
										"new_code": map[string]any{"type": "string"},
									},
									"required": []string{"path", "old_code", "new_code"},
								},
							},
							"add_code_files": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"path":    map[string]any{"type": "string"},
										"content": map[string]any{"type": "string"},
									},
									"required": []string{"path", "content"},
								},
							},
							"remove_code_files": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"path": map[string]any{"type": "string"},
									},
									"required": []string{"path"},
								},
							},
						},
						"required": []string{"markdown_summary", "code_updates", "add_code_files", "remove_code_files"},
					}),
				},
			},
		},
	},
	ToolChoice: &types.ToolChoiceMemberAny{
		Value: types.AnyToolChoice{},
	},
}
