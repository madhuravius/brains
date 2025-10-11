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

func TestRepairJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "balanced JSON stays same",
			in:   `{"a":1}`,
			want: `{"a":1}`,
		},
		{
			name: "missing closing brace",
			in:   `{"a":1`,
			want: `{"a":1}`,
		},
		{
			name: "missing closing bracket",
			in:   `[1, 2, 3`,
			want: `[1, 2, 3]`,
		},
		{
			name: "missing both",
			in:   `[{}`,
			want: `[{}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.RepairJSON(tt.in)
			if got != tt.want {
				t.Errorf("repairJSON(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractAnyJSON_SimpleObject(t *testing.T) {
	type Obj struct {
		A int `json:"a"`
	}

	raw := `junk before {"a":42} junk after`
	out, err := core.ExtractAnyJSON[Obj](raw)
	if err != nil {
		t.Fatalf("expected valid JSON extraction, got error: %v", err)
	}
	if out.A != 42 {
		t.Errorf("expected A=42, got %d", out.A)
	}
}

func TestExtractAnyJSON_ArrayCorrupted(t *testing.T) {
	raw := `**Project[ 	{ 	"name":"data_extractor", 	"parameters":{ 		"research_actions":{ 			"files_requested":[ "README.md" ], 			"urls_recommended":[] 		} 	}  } ]`
	out, err := core.ExtractAnyJSON[core.ResearchModelResponseWithParameters](raw)
	if err != nil {
		t.Fatalf("expected valid JSON extraction, got error: %v", err)
	}
	if out.Name != "data_extractor" {
		t.Errorf("expected A=data_extractor, got %s", out.Name)
	}
}

func TestExtractAnyJSON_Array(t *testing.T) {
	type Obj struct {
		Name string `json:"name"`
	}
	raw := `I[{"name":"data_extractor"}]I`
	out, err := core.ExtractAnyJSON[[]Obj](raw)
	if err != nil {
		t.Fatalf("expected valid JSON extraction, got error: %v", err)
	}
	if len(*out) != 1 || (*out)[0].Name != "data_extractor" {
		t.Errorf("unexpected output: %+v", *out)
	}
}

func TestExtractAnyJSON_UnclosedJSON(t *testing.T) {
	type Obj struct {
		Name string `json:"name"`
	}

	raw := `[{"name": "partial"`
	out, err := core.ExtractAnyJSON[[]Obj](raw)
	if err != nil {
		t.Fatalf("expected repaired valid JSON, got error: %v", err)
	}
	if len(*out) != 1 || (*out)[0].Name != "partial" {
		t.Errorf("unexpected result: %+v", *out)
	}
}

func TestExtractAnyJSON_NoValidJSON(t *testing.T) {
	type Obj struct{ X int }
	raw := `no json here at all`
	_, err := core.ExtractAnyJSON[Obj](raw)
	if err == nil {
		t.Fatalf("expected error for invalid input, got nil")
	}
}

func TestExtractAnyJSON_MultipleCandidates(t *testing.T) {
	type Obj struct {
		Val int `json:"val"`
	}

	raw := `nonsense {"val":1} junk {"val":2}`
	out, err := core.ExtractAnyJSON[Obj](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Val != 1 {
		t.Errorf("expected first valid JSON to be extracted, got %d", out.Val)
	}
}
