package repo_map

import (
	"bytes"
	"context"
	"os"
	"strings"

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
			if nameNode := n.ChildByFieldName("name"); nameNode != nil {
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
	appendSymbol := func(symType, name string, n *sitter.Node, doc string, params []Param) {
		symbols = append(symbols, &SymbolMap{
			Type:   symType,
			Name:   name,
			Start:  int(n.StartByte()),
			End:    int(n.EndByte()),
			Doc:    doc,
			Params: params,
		})
	}

	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
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
				doc := extractDoc(lang, spec, source)
				appendSymbol(symType, name, n, doc, nil)
			}
		}

		if lang == "python" && n.Type() == "assignment" {
			name := childContent(n, "left", source)
			if name != "" && name == strings.ToUpper(name) {
				doc := extractDoc(lang, n, source)
				appendSymbol("constant", name, n, doc, nil)
			}
		}

		if lang == "elixir" && n.Type() == "call" {
			if symType, name, params, ok := elixirSymbolFromCall(n, source); ok && name != "" {
				doc := extractElixirDoc(symType, n, source)
				appendSymbol(symType, name, n, doc, params)
			}
		}

		for _, rule := range rules {
			if n.Type() == rule.NodeType {
				name := ""
				if rule.FieldName != "" {
					name = childContent(n, rule.FieldName, source)
				} else {
					for i := 0; i < int(n.ChildCount()); i++ {
						if c := n.Child(i); c.Type() == "identifier" {
							name = nodeContent(c, source)
							break
						}
					}
				}
				if name != "" {
					doc := extractDoc(lang, n, source)
					var params []Param
					switch rule.SymbolType {
					case "function", "method", "constructor":
						params = extractParams(lang, n, source)
					}
					appendSymbol(rule.SymbolType, name, n, doc, params)
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

func childContent(n *sitter.Node, field string, src []byte) string {
	if c := n.ChildByFieldName(field); c != nil {
		return nodeContent(c, src)
	}
	return ""
}

func nodeContent(n *sitter.Node, src []byte) string {
	return string(bytes.TrimSpace(src[n.StartByte():n.EndByte()]))
}

func extractDoc(lang string, n *sitter.Node, src []byte) string {
	switch lang {
	case "python":
		if n.Type() == "function_definition" || n.Type() == "class_definition" || n.Type() == "module" {
			if s := pythonDocstring(n, src); s != "" {
				return s
			}
		}
		return leadingDocComments(lang, n, src)
	case "go":
		return leadingDocComments(lang, n, src)
	case "javascript", "typescript", "java", "csharp", "php", "cpp", "rust", "ruby":
		if s := nearestPrecedingDoc(lang, n, src); s != "" {
			return s
		}
		return leadingDocComments(lang, n, src)
	default:
		return ""
	}
}

func pythonDocstring(n *sitter.Node, src []byte) string {
	body := n.ChildByFieldName("body")
	if body == nil {
		for i := 0; i < int(n.ChildCount()); i++ {
			if n.Child(i).Type() == "block" {
				body = n.Child(i)
				break
			}
		}
	}
	if body == nil || body.ChildCount() == 0 {
		return ""
	}
	first := firstNonCommentChild(body)
	if first == nil || first.Type() != "expression_statement" || first.ChildCount() == 0 {
		return ""
	}
	if str := first.Child(0); str != nil && str.Type() == "string" {
		return stripPythonString(nodeContent(str, src))
	}
	return ""
}

func firstNonCommentChild(n *sitter.Node) *sitter.Node {
	for i := 0; i < int(n.ChildCount()); i++ {
		if c := n.Child(i); c.Type() != "comment" {
			return c
		}
	}
	return nil
}

func leadingDocComments(lang string, n *sitter.Node, src []byte) string {
	var parts []string
	anchor := docAnchor(n)
	if anchor == nil {
		return ""
	}
	start := anchor.StartByte()

	for ps := anchor.PrevSibling(); ps != nil; ps = ps.PrevSibling() {
		if !isCommentNode(lang, ps.Type()) {
			// Stop at the first non-comment sibling (prevents bleeding across items)
			break
		}
		txt := nodeContent(ps, src)
		// Require doc-style where appropriate
		if !isDocStyle(lang, txt) && lang != "go" && lang != "python" && lang != "ruby" {
			break
		}
		parts = append([]string{stripCommentMarkers(lang, txt)}, parts...)
		if hasBlankLineBetween(src, ps.StartByte(), start) {
			break
		}
		start = ps.StartByte()
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func isCommentNode(_ string, t string) bool {
	switch t {
	case "comment", "line_comment", "block_comment", "attribute_item":
		return true
	default:
		return false
	}
}

// docAnchor climbs to a wrapper statement that likely owns the leading doc comments.
// For JS/TS this includes export_statement; extend if you see other wrappers in your grammar.
func docAnchor(n *sitter.Node) *sitter.Node {
	if n == nil {
		return nil
	}
	isWrapper := func(t string) bool {
		switch t {
		case "export_statement", "decorated_definition":
			return true
		default:
			return false
		}
	}
	cur := n
	for p := cur.Parent(); p != nil; p = p.Parent() {
		if p.StartByte() == cur.StartByte() || isWrapper(p.Type()) {
			cur = p
			continue
		}
		break
	}
	return cur
}

func isDocStyle(lang, s string) bool {
	s = strings.TrimSpace(s)
	switch lang {
	case "javascript", "typescript", "java", "php":
		return strings.HasPrefix(s, "/**")
	case "csharp":
		return strings.HasPrefix(s, "///") || strings.HasPrefix(s, "/**")
	case "rust":
		return strings.HasPrefix(s, "///") || strings.HasPrefix(s, "/**") || strings.HasPrefix(s, "#[doc")
	default:
		return true
	}
}

func nearestPrecedingDoc(lang string, n *sitter.Node, src []byte) string {
	start := int(n.StartByte())
	if start <= 0 {
		return ""
	}
	data := src[:start]

	// Walk back to the last non-whitespace char before the node.
	i := len(data) - 1
	for i >= 0 && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i--
	}
	if i < 0 {
		return ""
	}

	// If we are positioned at the end of a block comment "*/", capture the nearest "/** ... */" block.
	if i >= 1 && data[i] == '/' && data[i-1] == '*' {
		end := i + 1 // position after '/'
		// Find the last "/*" before end; prefer "/**" (doc-style)
		startIdx := bytes.LastIndex(data[:end], []byte("/**"))
		if startIdx < 0 {
			// fallback to any block comment if there is no "/**"
			startIdx = bytes.LastIndex(data[:end], []byte("/*"))
		}
		if startIdx >= 0 {
			block := string(data[startIdx:end])
			return strings.TrimSpace(stripCommentMarkers(lang, block))
		}
	}

	// Else collect contiguous '///' lines right above (Rust/C# and also tolerated for others).
	isLineDoc := func(line string) bool {
		l := strings.TrimSpace(line)
		switch lang {
		case "rust", "csharp":
			return strings.HasPrefix(l, "///")
		default:
			// Some projects use /// in cpp/java/js/php too; allow it as a fallback
			return strings.HasPrefix(l, "///")
		}
	}

	lines := []string{}
	for i >= 0 {
		// find start of current line
		j := i
		for j >= 0 && data[j] != '\n' && data[j] != '\r' {
			j--
		}
		// slice is safe even when j == -1 (becomes [0:...])
		line := string(data[j+1 : i+1])
		trim := strings.TrimSpace(line)

		if trim == "" {
			// stop once we reached a blank line above the doc block
			if len(lines) > 0 {
				break
			}
			// skip leading blanks immediately before the node
			i = j - 1
			continue
		}
		if !isLineDoc(line) {
			break
		}
		// strip the leading "///"
		lines = append([]string{strings.TrimSpace(strings.TrimPrefix(trim, "///"))}, lines...)
		// move to previous line
		i = j - 1
	}
	if len(lines) > 0 {
		return strings.TrimSpace(strings.Join(lines, " "))
	}
	return ""
}

func stripCommentMarkers(_ string, s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		l := strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(l, "///"):
			lines[i] = strings.TrimSpace(strings.TrimPrefix(l, "///"))
		case strings.HasPrefix(l, "//"):
			lines[i] = strings.TrimSpace(strings.TrimPrefix(l, "//"))
		case strings.HasPrefix(l, "#"):
			lines[i] = strings.TrimSpace(strings.TrimPrefix(l, "#"))
		case strings.HasPrefix(l, "/*"):
			l = strings.TrimSpace(strings.TrimPrefix(l, "/*"))
			l = strings.TrimSpace(strings.TrimSuffix(l, "*/"))
			// Strip leading "*" commonly used in block comments
			l = strings.TrimLeft(l, "*")
			lines[i] = strings.TrimSpace(l)
		default:
			lines[i] = ln
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func stripPythonString(s string) string {
	p := strings.ToLower(s)
	for {
		if len(p) > 0 && (p[0] == 'r' || p[0] == 'u' || p[0] == 'f' || p[0] == 'b') {
			p = p[1:]
			s = s[1:]
			continue
		}
		break
	}
	stripQuotes := func(ss, quote string) (string, bool) {
		if strings.HasPrefix(ss, quote) && strings.HasSuffix(ss, quote) && len(ss) >= 2*len(quote) {
			return ss[len(quote) : len(ss)-len(quote)], true
		}
		return ss, false
	}
	if t, ok := stripQuotes(s, `"""`); ok {
		return t
	}
	if t, ok := stripQuotes(s, `'''`); ok {
		return t
	}
	if t, ok := stripQuotes(s, `"`); ok {
		return t
	}
	if t, ok := stripQuotes(s, `'`); ok {
		return t
	}
	return s
}

func hasBlankLineBetween(src []byte, a, b uint32) bool {
	if b <= a {
		return false
	}
	return strings.Contains(string(src[a:b]), "\n\n")
}

func extractParams(lang string, n *sitter.Node, src []byte) []Param {
	switch lang {
	case "go":
		return goParams(n, src)
	case "python":
		return pythonParams(n, src)
	case "javascript", "typescript":
		return jsTsParams(n, src)
	case "java":
		return javaParams(n, src)
	case "cpp":
		return cppParams(n, src)
	case "csharp":
		return csharpParams(n, src)
	case "php":
		return phpParams(n, src)
	case "rust":
		return rustParams(n, src)
	case "ruby":
		return rubyParams(n, src)
	case "elixir":
		return elixirParams(n, src)
	default:
		return nil
	}
}

func cleanTypeColon(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, ":") {
		s = strings.TrimSpace(s[1:])
	}
	return s
}

func extractFirstStringLiteralFromText(txt string) string {
	s := strings.TrimSpace(txt)
	for _, q := range []string{`"""`, `'''`, `"`, `'`} {
		if i := strings.Index(s, q); i >= 0 {
			rest := s[i+len(q):]
			if j := strings.Index(rest, q); j >= 0 {
				return strings.TrimSpace(rest[:j]) // trim result
			}
		}
	}
	return ""
}

func childByFieldOrTypes(n *sitter.Node, field string, types ...string) *sitter.Node {
	if c := n.ChildByFieldName(field); c != nil {
		return c
	}
	return findChildByTypes(n, types...)
}

func findChildByTypes(n *sitter.Node, types ...string) *sitter.Node {
	if len(types) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(types))
	for _, t := range types {
		set[t] = struct{}{}
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		c := n.Child(i)
		if _, ok := set[c.Type()]; ok {
			return c
		}
	}
	return nil
}

func argsNode(n *sitter.Node) *sitter.Node {
	return childByFieldOrTypes(n, "arguments", "arguments")
}

func isPunctType(t string) bool {
	return t == "(" || t == ")" || t == ","
}
