package config_test

import (
	"os"
	"path/filepath"
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

func TestSetContextFromGlobInvalidPattern(t *testing.T) {
	b := &config.BrainsConfig{}
	_, err := b.SetContextFromGlob("[")
	assert.Error(t, err)
}

func TestSetContextFromGlobFileReadError(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(dirPath, 0o700)

	b := &config.BrainsConfig{}
	_, err := b.SetContextFromGlob(filepath.Join(dirPath, "*"))
	assert.Error(t, err)
}
