package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

func LoadGitignore(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			pterm.Warning.Println(".gitignore not found, continuing")
			return []string{}, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

func (c *CommonToolsConfig) IsIgnored(path string) bool {
	path = filepath.ToSlash(path)

	if strings.HasPrefix(path, ".git/") || path == ".git" {
		return true
	}

	for _, pat := range c.ignorePatterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}

		pat = filepath.ToSlash(pat)
		if strings.HasSuffix(pat, "/") {
			if strings.HasPrefix(path, pat) {
				return true
			}
			continue
		}

		base := filepath.Base(path)
		if _, found := defaultIgnoreNames[base]; found {
			return true
		}

		if matched, _ := filepath.Match(pat, path); matched {
			return true
		}

		basePath := filepath.Base(path)
		if matched, _ := filepath.Match(pat, basePath); matched {
			return true
		}

		if strings.HasSuffix(pat, "*") {
			prefix := strings.TrimSuffix(pat, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}
