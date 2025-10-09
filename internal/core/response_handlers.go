package core

import (
	"encoding/json"
	"fmt"
)

func UnwrapFunc[T any, TP interface{ GetParameters() T }]() func(TP) T {
	return func(v TP) T { return v.GetParameters() }
}

func (r CodeModelResponse) IsHydrated() bool {
	return r.MarkdownSummary != "" ||
		len(r.CodeUpdates) > 0 ||
		len(r.AddCodeFiles) > 0 ||
		len(r.RemoveCodeFiles) > 0
}

func (r ResearchModelResponse) IsHydrated() bool {
	return len(r.ResearchActions.UrlsRecommended) > 0
}

func (w CodeModelResponseWithParameters) GetParameters() CodeModelResponse {
	return w.Parameters
}

func (w ResearchModelResponseWithParameters) GetParameters() ResearchModelResponse {
	return w.Parameters
}

func ExtractResponse[T Hydratable, TP any](respBody []byte, unwrap func(TP) T) (*T, error) {
	var raw json.RawMessage
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	tryDecode := func(extract func() T) (*T, bool) {
		out := extract()
		if !out.IsHydrated() {
			return nil, false
		}
		return &out, true
	}

	if len(raw) > 0 && raw[0] == '[' {
		// Array of T
		if r, ok := tryDecode(func() T {
			var arr []T
			_ = json.Unmarshal(raw, &arr)
			if len(arr) > 0 {
				return arr[0]
			}
			var zero T
			return zero
		}); ok {
			return r, nil
		}

		// Array of TP
		if r, ok := tryDecode(func() T {
			var arr []TP
			_ = json.Unmarshal(raw, &arr)
			if len(arr) > 0 {
				return unwrap(arr[0])
			}
			var zero T
			return zero
		}); ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not interpret JSON array as known types")
	}

	// Single T
	if r, ok := tryDecode(func() T {
		var v T
		_ = json.Unmarshal(raw, &v)
		return v
	}); ok {
		return r, nil
	}

	// Single TP
	if r, ok := tryDecode(func() T {
		var v TP
		_ = json.Unmarshal(raw, &v)
		return unwrap(v)
	}); ok {
		return r, nil
	}

	return nil, fmt.Errorf("unrecognized or empty JSON structure")
}
