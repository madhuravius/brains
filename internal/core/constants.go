package core

import (
	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

//go:embed schemas/code.json
var codeJSONSchema string

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
				Name:        aws.String("code_update"),
				Description: aws.String("Generate code updates in a specific schema"),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: document.NewLazyDocument(codeJSONSchema),
				},
			},
		},
	},
}
