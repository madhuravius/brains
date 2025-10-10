package repo_map

import (
	"bytes"
	"context"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
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
	appendSymbol := func(symType, name string, n *sitter.Node) {
		symbols = append(symbols, &SymbolMap{
			Type:  symType,
			Name:  name,
			Start: int(n.StartByte()),
			End:   int(n.EndByte()),
		})
	}

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		// --- special handling for Go type_declaration (structs/interfaces) ---
		if lang == "go" && n.Type() == "type_declaration" {
			for i := 0; i < int(n.ChildCount()); i++ {
				spec := n.Child(i)
				if spec.Type() != "type_spec" {
					continue
				}
				nameNode := spec.ChildByFieldName("name")
				if nameNode == nil {
					continue
				}
				name := nodeContent(nameNode, source)
				typeNode := spec.ChildByFieldName("type")
				symType := "type"
				if typeNode != nil {
					switch typeNode.Type() {
					case "struct_type":
						symType = "struct"
					case "interface_type":
						symType = "interface"
					}
				}
				appendSymbol(symType, name, n)
			}
		}

		// --- special handling for Python constants ---
		if lang == "python" && n.Type() == "assignment" {
			name := childContent(n, "left", source)
			if name != "" && name == strings.ToUpper(name) {
				appendSymbol("constant", name, n)
			}
		}

		// --- generic rule application ---
		for _, rule := range rules {
			if n.Type() == rule.NodeType {
				name := ""
				if rule.FieldName != "" {
					name = childContent(n, rule.FieldName, source)
				} else {
					for i := 0; i < int(n.ChildCount()); i++ {
						c := n.Child(i)
						if c.Type() == "identifier" {
							name = nodeContent(c, source)
							break
						}
					}
				}

				if name != "" {
					appendSymbol(rule.SymbolType, name, n)
				}
			}
		}

		// --- recurse down ---
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}

	walk(root)
	return symbols
}

func childContent(n *sitter.Node, field string, src []byte) string {
	c := n.ChildByFieldName(field)
	if c == nil {
		return ""
	}
	return nodeContent(c, src)
}

func nodeContent(n *sitter.Node, src []byte) string {
	return string(bytes.TrimSpace(src[n.StartByte():n.EndByte()]))
}
