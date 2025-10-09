package core_test

import (
	"encoding/json"
	"testing"

	"github.com/madhuravius/brains/internal/core"

	"github.com/stretchr/testify/assert"
)

func TestExtractCodeModelResponse_ArrayOfCodeModelResponse(t *testing.T) {
	resp := []core.CodeModelResponse{{MarkdownSummary: "array simple"}}
	data, _ := json.Marshal(resp)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, "array simple", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_ArrayOfCodeModelResponseWithParameters(t *testing.T) {
	wrapper := []core.CodeModelResponseWithParameters{{
		Name:       "test",
		Parameters: core.CodeModelResponse{MarkdownSummary: "array with params"},
	}}
	data, _ := json.Marshal(wrapper)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, "array with params", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_CodeModelResponse(t *testing.T) {
	resp := core.CodeModelResponse{MarkdownSummary: "object simple"}
	data, _ := json.Marshal(resp)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, "object simple", cmr.MarkdownSummary)
}

func TestExtractCodeModelResponse_CodeModelResponseWithParameters(t *testing.T) {
	wrapper := core.CodeModelResponseWithParameters{
		Name:       "test",
		Parameters: core.CodeModelResponse{MarkdownSummary: "object with params"},
	}
	data, _ := json.Marshal(wrapper)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, "object with params", cmr.MarkdownSummary)
}

func TestExtractResearchModelResponse_ArrayOfResearchModelResponse(t *testing.T) {
	resp := []core.ResearchModelResponse{{ResearchActions: core.ResearchActions{UrlsRecommended: []string{"url1"}}}}
	data, _ := json.Marshal(resp)

	rmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.ResearchModelResponse, core.ResearchModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, []string{"url1"}, rmr.ResearchActions.UrlsRecommended)
}

func TestExtractResearchModelResponse_ArrayOfResearchModelResponseWithParameters(t *testing.T) {
	wrapper := []core.ResearchModelResponseWithParameters{{
		Name:       "test",
		Parameters: core.ResearchModelResponse{ResearchActions: core.ResearchActions{UrlsRecommended: []string{"url2"}}},
	}}
	data, _ := json.Marshal(wrapper)

	rmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.ResearchModelResponse, core.ResearchModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, []string{"url2"}, rmr.ResearchActions.UrlsRecommended)
}

func TestExtractResearchModelResponse_ResearchModelResponse(t *testing.T) {
	resp := core.ResearchModelResponse{ResearchActions: core.ResearchActions{UrlsRecommended: []string{"url3"}}}
	data, _ := json.Marshal(resp)

	rmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.ResearchModelResponse, core.ResearchModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, []string{"url3"}, rmr.ResearchActions.UrlsRecommended)
}

func TestExtractResearchModelResponse_ResearchModelResponseWithParameters(t *testing.T) {
	wrapper := core.ResearchModelResponseWithParameters{
		Name:       "test",
		Parameters: core.ResearchModelResponse{ResearchActions: core.ResearchActions{UrlsRecommended: []string{"url4"}}},
	}
	data, _ := json.Marshal(wrapper)

	rmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.ResearchModelResponse, core.ResearchModelResponseWithParameters]())
	assert.NoError(t, err)
	assert.Equal(t, []string{"url4"}, rmr.ResearchActions.UrlsRecommended)
}

func TestExtractResponse_Unrecognized(t *testing.T) {
	data := []byte(`{"unexpected":"value"}`)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.Nil(t, cmr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized or empty JSON structure")
}

func TestExtractResponse_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json`)

	cmr, err := core.ExtractResponse(data, core.UnwrapFunc[core.CodeModelResponse, core.CodeModelResponseWithParameters]())
	assert.Nil(t, cmr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}
