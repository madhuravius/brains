package tools

import (
	"bufio"
	"os"
	pathpkg "path"
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

	// Ignore .git dir anywhere
	if strings.Contains("/"+path+"/", "/.git/") || path == ".git" || strings.HasSuffix(path, "/.git") {
		return true
	}

	// Apply default ignore names by basename early
	if base := filepath.Base(path); base != "" {
		if _, found := defaultIgnoreNames[base]; found {
			return true
		}
	}

	for _, pat := range c.ignorePatterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}

		pat = filepath.ToSlash(pat)

		// Handle directory patterns
		if strings.HasSuffix(pat, "/") {
			dir := strings.TrimSuffix(pat, "/")
			if dir == "" {
				continue
			}

			anchored := strings.HasPrefix(dir, "/")
			if anchored {
				// anchored to repo root: "/build/" matches "build/..."
				dir = strings.TrimPrefix(dir, "/")
				if strings.HasPrefix(path, dir+"/") || path == dir {
					return true
				}
			} else if strings.Contains("/"+path+"/", "/"+dir+"/") || path == dir || strings.HasSuffix(path, "/"+dir) {
				// unanchored: "node_modules/" matches at any depth
				// ensure segment boundaries to avoid matching "my_node_modules"
				return true
			}
			continue
		}

		// If pattern has no slash, it matches basenames at any depth
		if !strings.Contains(pat, "/") {
			if matched, _ := pathpkg.Match(pat, filepath.Base(path)); matched {
				return true
			}
			continue
		}

		// Pattern with slashes:
		// - If it starts with "/", treat it as anchored to repo root.
		// - Otherwise, try to match at any depth by sliding over path segments.
		if ap, ok := strings.CutPrefix(pat, "/"); ok {
			if matched, _ := pathpkg.Match(ap, path); matched {
				return true
			}
			continue
		}

		// Try unanchored match at any depth
		rest := path
		for {
			if matched, _ := pathpkg.Match(pat, rest); matched {
				return true
			}
			i := strings.IndexByte(rest, '/')
			if i < 0 {
				break
			}
			rest = rest[i+1:]
		}

		// Fallbacks you already had
		if matched, _ := pathpkg.Match(pat, path); matched {
			return true
		}
		if matched, _ := pathpkg.Match(pat, filepath.Base(path)); matched {
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
