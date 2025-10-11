package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/config"
)

func TestLoadConfigWhenNoFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWD) }()
	_ = os.Chdir(tmpDir)

	homeDir := t.TempDir()
	_ = os.Setenv("HOME", homeDir)

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, config.DefaultConfig.AWSRegion, cfg.GetConfig().AWSRegion)
	assert.Equal(t, config.DefaultConfig.Model, cfg.GetConfig().Model)
	assert.Empty(t, cfg.GetConfig().Personas)

	expectedPath := filepath.Join(tmpDir, ".brains.yml")
	_, statErr := os.Stat(expectedPath)
	assert.NoError(t, statErr)
}

func TestGetPersonaInstructions(t *testing.T) {
	b := &config.BrainsConfig{
		Personas: map[string]string{
			"dev": "You are a helpful developer.",
		},
	}

	instr := b.GetPersonaInstructions("dev")
	expected := "Human: You are a helpful developer.\n\n"
	assert.Equal(t, expected, instr)

	empty := b.GetPersonaInstructions("nonexistent")
	assert.Empty(t, empty)
}
