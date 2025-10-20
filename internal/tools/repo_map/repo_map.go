package repo_map

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madhuravius/brains/internal/tools"
)

func (r *RepoMapConfig) GetFiles() []*FileMap {
	return r.Files
}

func (r *RepoMapConfig) GetFileCount() int {
	return len(r.Files)
}

func (r *RepoMapConfig) ToPrompt() string {
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

func (r *RepoMapConfig) BuildRepoMap(ctx context.Context, repoRoot string) error {
	r.Path = repoRoot

	ignoreList, err := tools.LoadGitignore(fmt.Sprintf("%s/.gitignore", repoRoot))
	if err != nil {
		return err
	}
	r.commonTools = tools.NewCommonToolsConfig(ignoreList)

	err = filepath.Walk(repoRoot, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || !isSourceFile(path) {
			return nil
		}

		if r.commonTools.IsIgnored(path) {
			return nil
		}

		lang := detectLanguage(path)
		fileMap, err := r.ParseFile(ctx, path, lang)
		if err != nil {
			return err
		}
		r.Files = append(r.Files, fileMap)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
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
	for ln := range strings.SplitSeq(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			return strings.Join(strings.Fields(ln), " ")
		}
	}
	return ""
}
