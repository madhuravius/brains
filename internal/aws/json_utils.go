package aws

import (
	"encoding/json"
	"fmt"
	"strings"
)

// extractJSON extracts the first JSON payload from a raw string.
func extractJSON(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty response body")
	}

	// Strip common markdown code fences.
	raw = strings.ReplaceAll(raw, "```json", "")
	raw = strings.ReplaceAll(raw, "```", "")
	raw = strings.TrimSpace(raw)

	// Locate the first JSON start character.
	start := strings.IndexAny(raw, "{[")
	if start == -1 {
		return "", fmt.Errorf("no JSON start found")
	}

	// Decode the JSON from the first start character.
	dec := json.NewDecoder(strings.NewReader(raw[start:]))
	var rawMsg json.RawMessage
	if err := dec.Decode(&rawMsg); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}
	extracted := string(rawMsg)

	// If the extracted JSON is an array, wrap it in an object with a
	// "code_updates" field so callers expecting a CodeModelResponse can
	// unmarshal it directly.
	if strings.HasPrefix(strings.TrimSpace(extracted), "[") {
		wrapped := fmt.Sprintf("{\"code_updates\":%s}", extracted)
		return wrapped, nil
	}

	return extracted, nil
}
