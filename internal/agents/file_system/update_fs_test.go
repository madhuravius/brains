package file_system_test

import (
	"os"
	"path/filepath"
	"testing"

	"atomicgo.dev/keyboard"
	"github.com/stretchr/testify/assert"

	"brains/internal/agents/file_system"
)

func TestUpdateFile_ReplacesContent(t *testing.T) {
	tmp := t.TempDir()
	origWD, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(origWD) }()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	path := "sample.txt"
	old := "hello"
	new := "goodbye"
	assert.NoError(t, os.WriteFile(path, []byte(old), 0o600))

	ok, err := fs.UpdateFile(path, old, new, false)
	assert.True(t, ok)
	assert.NoError(t, err)

	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, new, string(data))
}

func TestUpdateFile_InteractiveShowsDiff(t *testing.T) {
	tmp := t.TempDir()
	origWD, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(origWD) }()

	go func() {
		_ = keyboard.SimulateKeyPress('y')
	}()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	path := "diff.txt"
	old := "foo"
	new := "bar"
	_ = os.WriteFile(path, []byte(old), 0o600)

	output := captureStdout(func() {
		ok, err := fs.UpdateFile(path, old, new, true)
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	contents, _ := os.ReadFile(path)
	// The rendered diff contains the literal word "diff" and the changed lines.
	assert.Contains(t, output, "diff")
	assert.Contains(t, output, "foo")
	assert.Contains(t, output, "bar")
	assert.Contains(t, string(contents), "bar")
}

func TestUpdateFile_Skipped(t *testing.T) {
	tmp := t.TempDir()
	origWD, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(origWD) }()

	go func() {
		_ = keyboard.SimulateKeyPress('n')
	}()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	path := "diff.txt"
	old := "foo"
	new := "bar"
	_ = os.WriteFile(path, []byte(old), 0o600)

	contents, _ := os.ReadFile(path)

	ok, err := fs.UpdateFile(path, old, new, true)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "foo")
}

func TestUpdateFile_ErrorWhenParentIsFile(t *testing.T) {
	tmp := t.TempDir()
	origWD, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(origWD) }()

	// Create a file that will be used as a directory in the update call.
	parentFile := "parent"
	_ = os.WriteFile(parentFile, []byte("x"), 0o600)

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	ok, err := fs.UpdateFile(filepath.Join(parentFile, "child.txt"), "", "data", false)
	assert.False(t, ok)
	assert.Error(t, err)
}
