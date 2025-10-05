package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractJSON_EmptyObject(t *testing.T) {
	input := "{}"
	expected := "{}"

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestExtractJSON_WithReasoningAndTrailingNoise(t *testing.T) {
	input := "Here is some reasoning output that should be ignored {\"key\":\"value\"} and some extra noise"
	expected := "{\"key\":\"value\"}"

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestExtractJSON_NestedObject(t *testing.T) {
	input := "blah blah ```json\n{\"code_updates\":[{\"path\":\"file.go\",\"old_code\":\"old\",\"new_code\":\"new\"}]}\n``` some trailing text"
	expected := "{\"code_updates\":[{\"path\":\"file.go\",\"old_code\":\"old\",\"new_code\":\"new\"}]}"

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestExtractJSON_ComplexNestedStructure(t *testing.T) {
	input := `
	Here is the response:
	{
		"metadata": {
			"request_id": "abc-123",
			"tags": ["test", "json", "nested"]
		},
		"data": [
			{
				"id": 1,
				"attributes": {"name": "Item 1", "price": 9.99}
			},
			{
				"id": 2,
				"attributes": {"name": "Item 2", "price": 19.99}
			}
		]
	}
	Thanks!`
	expected := `{
		"metadata": {
			"request_id": "abc-123",
			"tags": ["test", "json", "nested"]
		},
		"data": [
			{
				"id": 1,
				"attributes": {"name": "Item 1", "price": 9.99}
			},
			{
				"id": 2,
				"attributes": {"name": "Item 2", "price": 19.99}
			}
		]
	}`

	got, err := extractJSON(input)
	assert.NoError(t, err)
	// The function trims whitespace, so we compare after normalising whitespace.
	assert.JSONEq(t, expected, got)
}

func TestExtractJSON_EscapedQuotesInsideString(t *testing.T) {
	input := `Result: {"message":"He said, \"Hello, world!\"","status":"ok"}`
	expected := `{"message":"He said, \"Hello, world!\"","status":"ok"}`

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestExtractJSON_CodeBlockWithoutLanguageSpecifier(t *testing.T) {
	input := "Here is the payload:\n```\n{\"simple\":true,\"list\":[1,2,3]}\n```"
	expected := `{"simple":true,"list":[1,2,3]}`

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestExtractJSON_MultipleJsonObjectsOnlyFirstExtracted(t *testing.T) {
	input := `First: {"a":1} some text {"b":2}`
	expected := `{"a":1}`

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
