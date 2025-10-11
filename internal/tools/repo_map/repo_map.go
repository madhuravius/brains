package repo_map

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *RepoMap) ToPrompt() string {
	var sb strings.Builder
	for _, f := range r.Files {
		sb.WriteString(fmt.Sprintf("### File: %s\n", f.Path))
		for _, sym := range f.Symbols {
			if len(sym.Params) > 0 {
				var parts []string
				for _, p := range sym.Params {
					switch {
					case p.Type != "" && p.Name != "":
						parts = append(parts, fmt.Sprintf("%s: %s", p.Name, p.Type))
					case p.Name != "":
						parts = append(parts, p.Name)
					case p.Type != "":
						parts = append(parts, p.Type)
					}
				}
				sb.WriteString(fmt.Sprintf("- %s %s(%s)\n", sym.Type, sym.Name, strings.Join(parts, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf("- %s %s\n", sym.Type, sym.Name))
			}
			if sym.Doc != "" {
				doc := normalizeDocForPrompt(sym.Doc)
				if doc != "" {
					sb.WriteString(fmt.Sprintf("  Doc: %s\n", doc))
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func BuildRepoMap(ctx context.Context, repoRoot string) (*RepoMap, error) {
	var repo RepoMap
	repo.Path = repoRoot

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || !isSourceFile(path) {
			return nil
		}

		lang := detectLanguage(path)
		fileMap, err := ParseFile(ctx, path, lang)
		if err != nil {
			return err
		}
		repo.Files = append(repo.Files, fileMap)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func isSourceFile(path string) bool {
	switch filepath.Ext(path) {
	case ".go", ".py", ".js", ".ts", ".java", ".rb", ".cs", ".rs", ".php", ".cpp", ".ex", ".exs":
		return true
	default:
		return false
	}
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".cpp", ".cc", ".cxx", ".hpp", ".h":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".ex", ".exs":
		return "elixir"
	default:
		return "unknown"
	}
}

// normalizeDocForPrompt returns the first non-empty line of s,
// trims it, and collapses internal whitespace to a single space.
// It also normalizes CRLF/CR newlines.
func normalizeDocForPrompt(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			// collapse sequences of whitespace to a single space
			return strings.Join(strings.Fields(ln), " ")
		}
	}
	return ""
}
