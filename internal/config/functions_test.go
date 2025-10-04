package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"brains/internal/config"
)

func TestGetPersonaInstructionsFormatting(t *testing.T) {
	b := &config.BrainsConfig{
		Personas: map[string]string{
			"tester": "Test persona text.",
		},
	}
	result := b.GetPersonaInstructions("tester")
	assert.Equal(t, "Human: Test persona text.\n\n", result)
}
