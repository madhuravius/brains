package file_system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/pterm/pterm"
)

var defaultIgnoreNames = map[string]struct{}{
	"package-lock.json": {},
	"yarn.lock":         {},
	"pnpm-lock.yaml":    {},
	"bun.lockb":         {},
	"go.sum":            {},
	"poetry.lock":       {},
	"Cargo.lock":        {},
	"Gemfile.lock":      {},
	"composer.lock":     {},
	"Pipfile.lock":      {},
	"mix.lock":          {},
	"Podfile.lock":      {},
	"package.json.lock": {},
	"flake.lock":        {},
	"requirements.txt":  {},
	"target":            {},
	"node_modules":      {},
	".venv":             {},
}

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

func (f *FileSystemConfig) IsIgnored(path string) bool {
	path = filepath.ToSlash(path)

	if strings.HasPrefix(path, ".git/") || path == ".git" {
		return true
	}

	for _, pat := range f.ignorePatterns {
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

func (f *FileSystemConfig) GetFileContents(path string) (string, error) {
	if f.IsIgnored(path) {
		return "", nil
	}

	info, err := os.Stat(path)
	if err != nil {
		pterm.Warning.Printfln("failed to stat %s: %v", path, err)
		return "", err
	}
	if info.IsDir() {
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		pterm.Warning.Printfln("failed to read %s: %v", path, err)
		return "", err
	}

	return string(data), nil
}

func (f *FileSystemConfig) SetContextFromGlob(pattern string) (string, error) {
	files, err := doublestar.Glob(os.DirFS("."), pattern)
	if err != nil {
		return "", fmt.Errorf("failed to expand glob: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files matched pattern %s", pattern)
	}

	contents := make(map[string]string)

	for _, fpath := range files {
		data, err := f.GetFileContents(fpath)
		if err != nil {
			continue
		}
		if data == "" {
			continue
		}
		pterm.Debug.Printfln("added file to context: %s", fpath)
		contents[fpath] = data
	}

	contentData, err := json.Marshal(contents)
	if err != nil {
		pterm.Error.Printfln("failed to marshal file json map: %v", err)
		return "", err
	}

	return string(contentData), nil
}
