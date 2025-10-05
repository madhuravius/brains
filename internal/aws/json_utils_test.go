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

func TestExtractJSON_ArrayWrappedInCodeUpdates(t *testing.T) {
	input := "Here is the payload:\n```json\n[{\"path\":\"file.go\",\"old_code\":\"old\",\"new_code\":\"new\"}]\n```"
	expected := "{\"code_updates\":[{\"path\":\"file.go\",\"old_code\":\"old\",\"new_code\":\"new\"}]}"

	got, err := extractJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
