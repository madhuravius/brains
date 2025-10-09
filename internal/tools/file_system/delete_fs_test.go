package file_system_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"brains/internal/tools/file_system"
)

func TestDeleteFile_DeletesAndShowsDiff(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	path := "some/nested/file.txt"
	content := "original content"
	assert.NoError(t, fs.CreateFile(path, content))

	out := captureStdout(func() {
		err = fs.DeleteFile(path)
	})
	assert.NoError(t, err)

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "file should be removed")

	assert.Contains(t, out, "diff") // diff header rendered by DeleteFile
}

func TestDeleteFile_ReadErrorRemovalFails(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	dir := "parent"
	assert.NoError(t, os.MkdirAll(dir, 0o750))
	// put a child file so the directory is non‑empty – os.Remove will fail.
	child := filepath.Join(dir, "child.txt")
	assert.NoError(t, os.WriteFile(child, []byte("x"), 0o600))

	out := captureStdout(func() {
		err = fs.DeleteFile(dir)
	})
	// Expect an error because the directory is not empty.
	assert.Error(t, err)
	// The diff rendering should be skipped because reading failed.
	assert.NotContains(t, out, "diff")
}

func TestDeleteFile_NonExistentFile(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	nonExist := "does/not/exist.txt"
	out := captureStdout(func() {
		err = fs.DeleteFile(nonExist)
	})
	assert.Error(t, err)
	// No diff should be printed because the file could not be read.
	assert.NotContains(t, out, "diff")
}
