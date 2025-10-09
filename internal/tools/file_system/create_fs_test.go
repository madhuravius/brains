package file_system_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/tools/file_system"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestCreateFileCreatesAndPrintsDiff(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	path := "nested/hello.txt"
	content := "world"

	output := captureStdout(func() {
		err = fs.CreateFile(path, content)
	})
	assert.NoError(t, err)

	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
	assert.Contains(t, output, "diff")

	// Overwrite and ensure diff is rendered again.
	output2 := captureStdout(func() {
		err = fs.CreateFile(path, "new")
	})
	assert.NoError(t, err)
	data, _ = os.ReadFile(path)
	assert.Equal(t, "new", string(data))
	assert.Contains(t, output2, "diff")
}

func TestCreateFileFailsWhenParentIsFile(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	parentFile := "parent"
	assert.NoError(t, os.WriteFile(parentFile, []byte("marker"), 0o600))

	fs, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	err = fs.CreateFile(filepath.Join(parentFile, "child.txt"), "content")
	assert.Error(t, err)
}
