package config

import (
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

func (b *BrainsConfig) GetPersonaInstructions(persona string) string {
	personaText, found := b.Personas[persona]
	if !found {
		return ""
	}
	pterm.Info.Printfln("user electing to leverage persona (%s) with text: %s", persona, personaText)
	return fmt.Sprintf("Human: %s\n\n", personaText)
}

func (b *BrainsConfig) isIgnored(path string) bool {
	path = filepath.ToSlash(path)

	if strings.HasPrefix(path, ".git/") || path == ".git" {
		return true
	}

	for _, pat := range b.ignorePatterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}

		// Normalize pattern
		pat = filepath.ToSlash(pat)

		// Folder patterns: "foo/" → prefix match
		if strings.HasSuffix(pat, "/") {
			if strings.HasPrefix(path, pat) {
				return true
			}
			continue
		}

		// exclude general lock/large files
		base := filepath.Base(path)
		if _, found := defaultIgnoreNames[base]; found {
			return true
		}

		// Match full path directly
		if matched, _ := filepath.Match(pat, path); matched {
			return true
		}

		// Match on filename only (like *.log)
		basePath := filepath.Base(path)
		if matched, _ := filepath.Match(pat, basePath); matched {
			return true
		}

		// Match prefix for patterns like ".aider*" (which may cover directories)
		if strings.HasSuffix(pat, "*") {
			prefix := strings.TrimSuffix(pat, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

func (b *BrainsConfig) SetContextFromGlob(pattern string) (string, error) {
	files, err := doublestar.Glob(os.DirFS("."), pattern)
	if err != nil {
		return "", fmt.Errorf("failed to expand glob: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files matched pattern %s", pattern)
	}

	contents := make(map[string]string)

	for _, fpath := range files {
		if b.isIgnored(fpath) {
			continue
		}

		info, err := os.Stat(fpath)
		if err != nil {
			pterm.Warning.Printfln("failed to stat %s: %v", fpath, err)
			continue
		}
		if info.IsDir() {
			continue
		}

		data, err := os.ReadFile(fpath)
		if err != nil {
			pterm.Warning.Printfln("failed to read %s: %v", fpath, err)
			continue
		}

		pterm.Info.Printfln("added file to context: %s", fpath)
		contents[fpath] = string(data)
	}

	contentData, err := json.Marshal(contents)
	if err != nil {
		pterm.Error.Printfln("failed to marshal file json map: %v", err)
		return "", err
	}

	return string(contentData), nil
}
