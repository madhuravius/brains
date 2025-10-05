package aws

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

	dec := json.NewDecoder(strings.NewReader(raw[start:]))
	var rawMsg json.RawMessage
	if err := dec.Decode(&rawMsg); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}
	return string(rawMsg), nil
}
