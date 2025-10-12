package file_system_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/tools/file_system"
)

func TestSetContextFromGlob(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")
	_ = os.WriteFile(file1, []byte("content A"), 0o600)
	_ = os.WriteFile(file2, []byte("content B"), 0o600)

	f, _ := file_system.NewFileSystemConfig()

	origWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWD) }()
	_ = os.Chdir(tmpDir)

	ctx, err := f.SetContextFromGlob("*.txt")
	assert.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal([]byte(ctx), &result)
	assert.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "content A", result["a.txt"])
	assert.Equal(t, "content B", result["b.txt"])
}

func TestSetContextFromGlobInvalidPattern(t *testing.T) {
	f, _ := file_system.NewFileSystemConfig()
	_, err := f.SetContextFromGlob("[")
	assert.Error(t, err)
}

func TestSetContextFromGlobFileReadError(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(dirPath, 0o700)

	f, _ := file_system.NewFileSystemConfig()
	_, err := f.SetContextFromGlob(filepath.Join(dirPath, "*"))
	assert.Error(t, err)
}
