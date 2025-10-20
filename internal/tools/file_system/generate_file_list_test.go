package file_system_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/tools/file_system"
)

func TestFileListSuccess(t *testing.T) {
	f, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	result, err := f.GetFileTree("./test_fixtures/project1")
	assert.NoError(t, err)

	assert.Equal(t, result, `. (root)
  1.txt
  2.txt
  test/
    nested-1.txt
    nested-2.txt
    nested-test/
      nested-nested-1.txt
      nested-nested-2.txt
`)
}

func TestFileListFailureNotExists(t *testing.T) {
	f, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	_, err = f.GetFileTree("./DOESNOTEXIST")
	assert.ErrorContains(t, err, "is not a directory")
}

func TestFileListFailureNotParent(t *testing.T) {
	f, err := file_system.NewFileSystemConfig()
	assert.NoError(t, err)

	_, err = f.GetFileTree("./test_fixtures/1.txt")
	assert.ErrorContains(t, err, "is not a directory")
}
