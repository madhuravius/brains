package core

import (
	"encoding/json"
	"fmt"
	"strings"
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
	return r.MarkdownSummary != "" || len(r.ResearchActions.UrlsRecommended) > 0 || len(r.ResearchActions.FilesRequested) > 0
}

func (w CodeModelResponseWithParameters) GetParameters() CodeModelResponse {
	return w.Parameters
}

func (w ResearchModelResponseWithParameters) GetParameters() ResearchModelResponse {
	return w.Parameters
}

func ExtractResponse[T Hydratable, TP any](
	respBody []byte,
	unwrap func(TP) T,
) (*T, error) {
	// Step 1: Try direct decode (fast path)
	var raw json.RawMessage
	if err := json.Unmarshal(respBody, &raw); err != nil {
		// Step 2: Fallback — attempt to extract recoverable JSON
		recovered, extErr := ExtractAnyJSON[json.RawMessage](string(respBody))
		if extErr != nil {
			return nil, fmt.Errorf("invalid JSON and no recoverable fragment: %w", err)
		}
		raw = *recovered
	}

	tryDecode := func(extract func() T) (*T, bool) {
		out := extract()
		if !out.IsHydrated() {
			return nil, false
		}
		return &out, true
	}

	// Step 3: Handle array-based responses
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

	// Step 4: Single object decode paths
	if r, ok := tryDecode(func() T {
		var v T
		_ = json.Unmarshal(raw, &v)
		return v
	}); ok {
		return r, nil
	}

	if r, ok := tryDecode(func() T {
		var v TP
		_ = json.Unmarshal(raw, &v)
		return unwrap(v)
	}); ok {
		return r, nil
	}

	return nil, fmt.Errorf("unrecognized or empty JSON structure")
}

func ExtractAnyJSON[T any](raw string) (*T, error) {
	raw = strings.TrimSpace(raw)

	// Find first possible JSON start
	start := strings.IndexAny(raw, "{[")
	if start == -1 {
		return nil, fmt.Errorf("no JSON start found")
	}
	raw = raw[start:]

	// Try to find where JSON ends by counting braces/brackets
	var (
		depth      int
		inString   bool
		escapeNext bool
		end        int
	)

	for i, ch := range raw {
		switch ch {
		case '\\':
			if inString {
				escapeNext = !escapeNext
			}
		case '"':
			if !escapeNext {
				inString = !inString
			}
			escapeNext = false
		case '{', '[':
			if !inString {
				depth++
			}
		case '}', ']':
			if !inString {
				depth--
				if depth == 0 {
					end = i + 1
					goto FOUND
				}
			}
		default:
			escapeNext = false
		}
	}

FOUND:
	candidate := raw
	if end > 0 {
		candidate = raw[:end]
	}

	candidate = RepairJSON(candidate)

	// 1. Try single object
	var v T
	if err := json.Unmarshal([]byte(candidate), &v); err == nil {
		return &v, nil
	}

	// 2. Try array of objects
	var arr []T
	if err := json.Unmarshal([]byte(candidate), &arr); err == nil && len(arr) > 0 {
		return &arr[0], nil
	}

	return nil, fmt.Errorf("no valid JSON found in input")
}

func RepairJSON(s string) string {
	openBraces := strings.Count(s, "{")
	closeBraces := strings.Count(s, "}")
	openBrackets := strings.Count(s, "[")
	closeBrackets := strings.Count(s, "]")

	for closeBraces < openBraces {
		s += "}"
		closeBraces++
	}
	for closeBrackets < openBrackets {
		s += "]"
		closeBrackets++
	}
	return s
}
