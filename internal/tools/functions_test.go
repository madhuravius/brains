package tools_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/tools"
)

func TestIsIgnoredDefaultAndGitignorePatterns(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()

	_ = os.Chdir(tmp)
	_ = os.WriteFile(".gitignore", []byte("*.tmp\nlogs/"), 0o600)

	patterns, err := tools.LoadGitignore(".gitignore")
	assert.NoError(t, err)
	assert.NotEmpty(t, patterns)

	cfg := tools.NewCommonToolsConfig(patterns)

	assert.True(t, cfg.IsIgnored(".git/config"))
	assert.True(t, cfg.IsIgnored("package-lock.json"))

	assert.True(t, cfg.IsIgnored("debug.tmp"))

	assert.False(t, cfg.IsIgnored("main.go"))
	assert.False(t, cfg.IsIgnored("src/utils/helper.go"))
}

func TestLoadGitignoreFileNotFound(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(tmp)

	patterns, err := tools.LoadGitignore(".gitignore")
	assert.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestLoadGitignoreParsesPatterns(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	_ = os.Chdir(tmp)

	content := "# comment line\n*.log\nbuild/\n"
	_ = os.WriteFile(".gitignore", []byte(content), 0o600)

	patterns, err := tools.LoadGitignore(".gitignore")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"*.log", "build/"}, patterns)
}
