package config_test

import (
	"encoding/json"
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

	// Change working directory to the temporary directory so that the glob works.
	origWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWD) }()
	_ = os.Chdir(tmpDir)

	ctx, err := b.SetContextFromGlob("*.txt")
	assert.NoError(t, err)

	// The function returns a JSON string mapping file names to their contents.
	var result map[string]string
	err = json.Unmarshal([]byte(ctx), &result)
	assert.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "content A", result["a.txt"])
	assert.Equal(t, "content B", result["b.txt"])
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
