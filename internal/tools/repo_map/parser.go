package repo_map

import (
	"context"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func ParseFile(ctx context.Context, path, lang string) (*FileMap, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := sitter.NewParser()

	switch lang {
	case "go":
		parser.SetLanguage(golang.GetLanguage())
	case "python":
		parser.SetLanguage(python.GetLanguage())
	case "javascript":
		parser.SetLanguage(javascript.GetLanguage())
	case "typescript":
		parser.SetLanguage(typescript.GetLanguage())
	case "java":
		parser.SetLanguage(java.GetLanguage())
	case "cpp":
		parser.SetLanguage(cpp.GetLanguage())
	case "csharp":
		parser.SetLanguage(csharp.GetLanguage())
	case "ruby":
		parser.SetLanguage(ruby.GetLanguage())
	case "php":
		parser.SetLanguage(php.GetLanguage())
	case "rust":
		parser.SetLanguage(rust.GetLanguage())
	case "elixir":
		parser.SetLanguage(elixir.GetLanguage())
	default:
		return &FileMap{Path: path, Language: lang}, nil
	}

	tree, err := parser.ParseCtx(ctx, nil, code)
	if err != nil {
		return nil, err
	}
	root := tree.RootNode()

	symbols := ExtractSymbols(lang, root, code)

	return &FileMap{
		Path:     path,
		Language: lang,
		Symbols:  symbols,
	}, nil
}

func ExtractFunctions(root *sitter.Node, source []byte) []string {
	var funcs []string

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "function_declaration" {
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				funcs = append(funcs, nameNode.Content(source))
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return funcs
}

func ExtractSymbols(lang string, root *sitter.Node, source []byte) []*SymbolMap {
	rules, ok := LanguageSymbolRules[lang]
	if !ok {
		return nil
	}

	var symbols []*SymbolMap
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		for _, rule := range rules {
			if n.Type() == rule.NodeType {
				name := childContent(n, rule.FieldName, source)
				if name != "" {
					symbols = append(symbols, &SymbolMap{
						Type:  rule.SymbolType,
						Name:  name,
						Start: int(n.StartByte()),
						End:   int(n.EndByte()),
					})
				}
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return symbols
}

func childContent(n *sitter.Node, field string, source []byte) string {
	child := n.ChildByFieldName(field)
	if child == nil {
		return ""
	}
	return child.Content(source)
}
