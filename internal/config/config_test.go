package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"brains/internal/config"
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
	assert.Equal(t, config.DefaultConfig.AWSRegion, cfg.AWSRegion)
	assert.Equal(t, config.DefaultConfig.Model, cfg.Model)
	assert.Empty(t, cfg.Personas)

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

func TestSetContextFromGlob(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")
	_ = os.WriteFile(file1, []byte("content A"), 0o600)
	_ = os.WriteFile(file2, []byte("content B"), 0o600)

	b := &config.BrainsConfig{}
	ctx, err := b.SetContextFromGlob(filepath.Join(tmpDir, "*.txt"))
	assert.NoError(t, err)

	assert.Contains(t, ctx, "--- a.txt ---")
	assert.Contains(t, ctx, "content A")
	assert.Contains(t, ctx, "--- b.txt ---")
	assert.Contains(t, ctx, "content B")
}
