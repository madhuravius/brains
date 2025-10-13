package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/config"
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

func TestPreCommandsSuccess(t *testing.T) {
	b := &config.BrainsConfig{
		PreCommands: []string{"exit 0"},
	}
	err := b.PreCommandsHook()
	assert.Nil(t, err)
}

func TestPreCommandsFailure(t *testing.T) {
	b := &config.BrainsConfig{
		PreCommands: []string{"exit 1"},
	}
	err := b.PreCommandsHook()
	assert.NotNil(t, err)
}
