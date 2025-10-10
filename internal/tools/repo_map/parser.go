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

func hasNonWhitespaceBetween(src []byte, a, b uint32) bool {
	if b <= a {
		return false
	}
	for _, r := range src[a:b] {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
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

func cppParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameters", "parameter_list")
	if plist == nil {
		plist = findFirstByType(n, "parameter_list")
	}
	if plist == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(plist.ChildCount()); i++ {
		p := plist.Child(i)
		switch p.Type() {
		case "parameter_declaration":
			typ := childContent(p, "type", src)
			name := childContent(p, "declarator", src)
			if name == "" {
				if d := p.ChildByFieldName("declarator"); d != nil {
					name = firstIdentifier(d, src)
				} else {
					name = firstIdentifier(p, src)
				}
			}
			def := childContent(p, "default_value", src)
			out = append(out, Param{Name: strings.TrimSpace(name), Type: strings.TrimSpace(typ), Default: strings.TrimSpace(def)})
		case "variadic_parameter":
			out = append(out, Param{Name: "...", Type: ""})
		}
	}
	return out
}

func csharpParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameter_list", "parameter_list")
	if plist == nil {
		return nil
	}
	var out []Param
	hasParams := false

	for i := 0; i < int(plist.ChildCount()); i++ {
		p := plist.Child(i)
		t := p.Type()
		if t != "parameter" && t != "this_parameter" && t != "parameter_array" && t != "params_parameter" {
			continue
		}
		typ := childContent(p, "type", src)
		name := childContent(p, "name", src)
		if name == "" {
			name = firstIdentifier(p, src)
		}
		def := childContent(p, "default_value", src)

		txt := nodeContent(p, src)
		prefix := ""
		if strings.Contains(txt, "this ") {
			prefix = "this "
		}
		if strings.Contains(txt, "ref ") {
			prefix += "ref "
		}
		if strings.Contains(txt, "out ") {
			prefix += "out "
		}
		if strings.Contains(txt, "in ") && !strings.Contains(txt, "using") {
			prefix += "in "
		}
		if strings.Contains(txt, "params ") || t == "parameter_array" || t == "params_parameter" {
			prefix += "params "
			hasParams = true
		}
		if prefix != "" && name != "" {
			name = strings.TrimSpace(prefix) + " " + name
		}
		out = append(out, Param{Name: strings.TrimSpace(name), Type: strings.TrimSpace(typ), Default: strings.TrimSpace(def)})
	}

	// Fallback: if signature text contains "params " but we didn't capture it (grammar oddity),
	// append a synthetic params arg to surface "params " in prompt.
	if !hasParams {
		txt := nodeContent(plist, src)
		if strings.Contains(txt, "params ") {
			// Try to extract the last word as a name (best-effort)
			tail := txt[strings.LastIndex(txt, "params ")+len("params "):]
			words := strings.Fields(tail)
			name := "rest"
			if len(words) > 0 {
				w := words[len(words)-1]
				name = strings.TrimRight(w, ",)")
			}
			out = append(out, Param{Name: "params " + name})
		}
	}
	return out
}

func phpParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameters", "parameters", "formal_parameters")
	if plist == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(plist.ChildCount()); i++ {
		p := plist.Child(i)
		ty := p.Type()
		if ty != "parameter" && ty != "simple_parameter" && ty != "variadic_parameter" {
			continue
		}
		typ := childContent(p, "type", src)
		name := childContent(p, "name", src)
		if name == "" {
			if vn := findFirstByType(p, "variable_name"); vn != nil {
				name = nodeContent(vn, src)
			} else {
				name = firstIdentifier(p, src)
			}
		}
		def := childContent(p, "default_value", src)

		txt := nodeContent(p, src)
		if strings.Contains(txt, "&") && !strings.HasPrefix(strings.TrimSpace(name), "&") {
			name = "&" + strings.TrimSpace(name)
		}
		if ty == "variadic_parameter" || strings.Contains(txt, "...") {
			if !strings.HasPrefix(strings.TrimSpace(name), "...") {
				name = "..." + strings.TrimSpace(name)
			}
		}
		out = append(out, Param{Name: name, Type: strings.TrimSpace(typ), Default: strings.TrimSpace(def)})
	}
	return out
}

func rustParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameters", "parameters")
	if plist == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(plist.ChildCount()); i++ {
		p := plist.Child(i)
		switch p.Type() {
		case "self_parameter":
			selfTxt := strings.ReplaceAll(nodeContent(p, src), " ", "")
			switch selfTxt {
			case "self":
				out = append(out, Param{Name: "self"})
			case "&self":
				out = append(out, Param{Name: "&self"})
			case "&mutself", "mutself", "self:mut":
				out = append(out, Param{Name: "&mut self"})
			default:
				out = append(out, Param{Name: nodeContent(p, src)})
			}
		case "parameter":
			name := childContent(p, "pattern", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			typ := childContent(p, "type", src)
			out = append(out, Param{Name: strings.TrimSpace(name), Type: strings.TrimSpace(typ)})
		}
	}
	return out
}

func rubyParams(n *sitter.Node, src []byte) []Param {
	var paramsNode *sitter.Node
	if pn := n.ChildByFieldName("parameters"); pn != nil {
		paramsNode = pn
	} else {
		var dfs func(*sitter.Node)
		dfs = func(nn *sitter.Node) {
			if paramsNode != nil {
				return
			}
			if nn.Type() == "parameters" {
				paramsNode = nn
				return
			}
			for i := 0; i < int(nn.ChildCount()); i++ {
				dfs(nn.Child(i))
				if paramsNode != nil {
					return
				}
			}
		}
		dfs(n)
	}
	if paramsNode == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		p := paramsNode.Child(i)
		switch p.Type() {
		case "required_parameter":
			name := childContent(p, "name", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			out = append(out, Param{Name: name})
		case "optional_parameter":
			name := childContent(p, "name", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			def := childContent(p, "value", src)
			if def == "" {
				def = childContent(p, "default", src)
			}
			out = append(out, Param{Name: name, Default: def})
		case "splat_parameter", "rest_parameter":
			name := firstIdentifier(p, src)
			if name != "" && !strings.HasPrefix(name, "*") {
				name = "*" + name
			} else if name == "" {
				name = "*"
			}
			out = append(out, Param{Name: name})
		case "hash_splat_parameter", "keyword_splat_parameter":
			name := firstIdentifier(p, src)
			if name != "" && !strings.HasPrefix(name, "**") {
				name = "**" + name
			} else if name == "" {
				name = "**"
			}
			out = append(out, Param{Name: name})
		case "keyword_parameter":
			name := childContent(p, "name", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			def := childContent(p, "value", src)
			out = append(out, Param{Name: name + ":", Default: def})
		case "block_parameter":
			name := firstIdentifier(p, src)
			if name != "" && !strings.HasPrefix(name, "&") {
				name = "&" + name
			} else if name == "" {
				name = "&"
			}
			out = append(out, Param{Name: name})
		case "forwarding_parameter":
			out = append(out, Param{Name: "..."})
		default:
			if p.Type() == "identifier" {
				out = append(out, Param{Name: nodeContent(p, src)})
			}
		}
	}
	return out
}

func goParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameters", "parameter_list")
	if plist == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(plist.ChildCount()); i++ {
		c := plist.Child(i)
		if c.Type() != "parameter_declaration" {
			continue
		}
		var names []string
		var typ string
		if tnode := c.ChildByFieldName("type"); tnode != nil {
			typ = nodeContent(tnode, src)
		}
		for j := 0; j < int(c.ChildCount()); j++ {
			if ch := c.Child(j); ch.Type() == "identifier" {
				names = append(names, nodeContent(ch, src))
			}
		}
		if len(names) == 0 {
			out = append(out, Param{Name: "", Type: typ})
		} else {
			for _, nm := range names {
				out = append(out, Param{Name: nm, Type: typ})
			}
		}
	}
	return out
}

func pythonParams(n *sitter.Node, src []byte) []Param {
	params := childByFieldOrTypes(n, "parameters", "parameters")
	if params == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(params.ChildCount()); i++ {
		p := params.Child(i)
		switch p.Type() {
		case "typed_parameter", "parameter":
			name := childContent(p, "name", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			typ := childContent(p, "type", src)
			def := childContent(p, "default", src)
			out = append(out, Param{Name: name, Type: typ, Default: def})
		case "default_parameter":
			name := childContent(p, "name", src)
			if name == "" {
				name = firstIdentifier(p, src)
			}
			typ := childContent(p, "type", src)
			def := childContent(p, "value", src)
			if def == "" {
				def = childContent(p, "default", src)
			}
			out = append(out, Param{Name: name, Type: typ, Default: def})
		case "list_splat_pattern":
			name := firstIdentifier(p, src)
			out = append(out, Param{Name: "*" + name})
		case "dictionary_splat_pattern":
			name := firstIdentifier(p, src)
			out = append(out, Param{Name: "**" + name})
		case "identifier":
			out = append(out, Param{Name: nodeContent(p, src)})
		}
	}
	return out
}

func jsTsParams(n *sitter.Node, src []byte) []Param {
	// Find params anywhere under this node (TS: formal_parameters, JS: parameters/formal_parameters)
	paramsNode := findFirstByType(n, "formal_parameters")
	if paramsNode == nil {
		paramsNode = findFirstByType(n, "parameters")
	}
	if paramsNode == nil {
		// Fallback textual parse
		return jsLikeFallbackParams(nodeContent(n, src))
	}

	var out []Param
	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		p := paramsNode.Child(i)
		typ := p.Type()
		switch typ {
		case "required_parameter", "optional_parameter", "parameter", "rest_parameter":
			name := firstIdentifier(p, src)
			typText := childContent(p, "type", src)
			if typText == "" {
				if tan := findFirstByType(p, "type_annotation"); tan != nil {
					typText = nodeContent(tan, src)
				}
			}
			typText = cleanTypeColon(typText)

			txt := nodeContent(p, src)
			if typ == "rest_parameter" || strings.Contains(txt, "...") {
				if name != "" && !strings.HasPrefix(name, "...") {
					name = "..." + name
				}
			}
			out = append(out, Param{Name: name, Type: typText})

		// JS-specific: parameters may be bare identifiers or assignment patterns
		case "identifier":
			out = append(out, Param{Name: nodeContent(p, src)})
		case "assignment_pattern":
			out = append(out, Param{Name: firstIdentifier(p, src)})
		}
	}
	// If we still found nothing (rare), try textual fallback
	if len(out) == 0 {
		return jsLikeFallbackParams(nodeContent(n, src))
	}
	return out
}

func cleanTypeColon(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, ":") {
		s = strings.TrimSpace(s[1:])
	}
	return s
}

func jsLikeFallbackParams(text string) []Param {
	start := strings.Index(text, "(")
	end := strings.LastIndex(text, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return nil
	}
	inner := text[start+1 : end]
	parts := strings.Split(inner, ",")
	var out []Param
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name := p
		if idx := strings.Index(p, ":"); idx >= 0 {
			name = strings.TrimSpace(p[:idx])
		}
		if strings.Contains(p, "...") && !strings.HasPrefix(name, "...") {
			name = "..." + strings.TrimPrefix(name, "...")
		}
		// Reduce to the first tokenish chunk
		fields := strings.FieldsFunc(name, func(r rune) bool {
			return r == ',' || r == '='
		})
		if len(fields) > 0 {
			name = fields[0]
		}
		out = append(out, Param{Name: name})
	}
	return out
}

func javaParams(n *sitter.Node, src []byte) []Param {
	fp := childByFieldOrTypes(n, "parameters", "formal_parameters")
	if fp == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(fp.ChildCount()); i++ {
		c := fp.Child(i)
		if c.Type() != "formal_parameter" && c.Type() != "receiver_parameter" && c.Type() != "spread_parameter" {
			continue
		}
		typ := childContent(c, "type", src)
		name := childContent(c, "name", src)
		if name == "" {
			name = firstIdentifier(c, src)
		}
		if c.Type() == "spread_parameter" && name != "" {
			name = "..." + name
		}
		out = append(out, Param{Name: name, Type: typ})
	}
	return out
}

func elixirParams(n *sitter.Node, src []byte) []Param {
	switch elixirCallName(n, src) {
	case "def", "defp", "defmacro", "defmacrop":
		if head := elixirFindHeadCall(n, src); head != nil {
			return elixirHeadParams(head, src)
		}
	}
	return nil
}

func firstIdentifier(n *sitter.Node, src []byte) string {
	var res string
	var dfs func(*sitter.Node)
	dfs = func(nn *sitter.Node) {
		if res != "" {
			return
		}
		if nn.Type() == "identifier" {
			res = nodeContent(nn, src)
			return
		}
		for i := 0; i < int(nn.ChildCount()); i++ {
			dfs(nn.Child(i))
			if res != "" {
				return
			}
		}
	}
	dfs(n)
	return res
}

func findFirstByType(n *sitter.Node, t string) *sitter.Node {
	if n.Type() == t {
		return n
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		if r := findFirstByType(n.Child(i), t); r != nil {
			return r
		}
	}
	return nil
}

func elixirSymbolFromCall(n *sitter.Node, src []byte) (symType, name string, params []Param, ok bool) {
	kw := elixirCallName(n, src)
	switch kw {
	case "defmodule":
		symType = "module"
		name = elixirFirstArgText(n, src)
	case "defprotocol":
		symType = "protocol"
		name = elixirFirstArgText(n, src)
	case "defimpl":
		symType = "impl"
		name = elixirFirstArgText(n, src)
	case "def", "defp":
		symType = "function"
		if head := elixirFindHeadCall(n, src); head != nil {
			name = elixirHeadName(head, src)
			params = elixirHeadParams(head, src)
		}
	case "defmacro", "defmacrop":
		symType = "macro"
		if head := elixirFindHeadCall(n, src); head != nil {
			name = elixirHeadName(head, src)
			params = elixirHeadParams(head, src)
		}
	default:
		return "", "", nil, false
	}
	return symType, name, params, name != ""
}

func elixirCallName(n *sitter.Node, src []byte) string {
	if f := n.ChildByFieldName("function"); f != nil {
		if name := elixirExtractFunctionName(f, src); name != "" {
			return name
		}
	}
	if f := n.ChildByFieldName("callee"); f != nil {
		if name := elixirExtractFunctionName(f, src); name != "" {
			return name
		}
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		switch c := n.Child(i); c.Type() {
		case "identifier", "dot", "qualified_call", "call":
			if name := elixirExtractFunctionName(c, src); name != "" {
				return name
			}
		}
	}
	txt := strings.TrimSpace(nodeContent(n, src))
	start := 0
	for start < len(txt) && (txt[start] == ' ' || txt[start] == '\t' || txt[start] == '\n' || txt[start] == '(') {
		start++
	}
	for i := start; i < len(txt); i++ {
		switch txt[i] {
		case ' ', '\t', '\n', '(':
			return strings.TrimSpace(txt[start:i])
		}
	}
	if start < len(txt) {
		return strings.TrimSpace(txt[start:])
	}
	return ""
}

func elixirExtractFunctionName(fn *sitter.Node, src []byte) string {
	switch fn.Type() {
	case "identifier":
		return nodeContent(fn, src)
	case "dot":
		if r := fn.ChildByFieldName("right"); r != nil && r.Type() == "identifier" {
			return nodeContent(r, src)
		}
		for i := int(fn.ChildCount()) - 1; i >= 0; i-- {
			if c := fn.Child(i); c.Type() == "identifier" {
				return nodeContent(c, src)
			}
		}
	case "qualified_call", "call":
		if f := fn.ChildByFieldName("function"); f != nil {
			if name := elixirExtractFunctionName(f, src); name != "" {
				return name
			}
		}
		if f := fn.ChildByFieldName("callee"); f != nil {
			if name := elixirExtractFunctionName(f, src); name != "" {
				return name
			}
		}
		for i := 0; i < int(fn.ChildCount()); i++ {
			if name := elixirExtractFunctionName(fn.Child(i), src); name != "" {
				return name
			}
		}
	}
	return ""
}

func elixirFirstArgText(n *sitter.Node, src []byte) string {
	args := argsNode(n)
	if args == nil || args.ChildCount() == 0 {
		return ""
	}
	for i := 0; i < int(args.ChildCount()); i++ {
		c := args.Child(i)
		if isPunctType(c.Type()) {
			continue
		}
		return strings.TrimSpace(nodeContent(c, src))
	}
	return ""
}

func elixirFindHeadCall(defCall *sitter.Node, src []byte) *sitter.Node {
	args := argsNode(defCall)
	if args == nil {
		return nil
	}
	for i := 0; i < int(args.ChildCount()); i++ {
		if c := args.Child(i); c.Type() == "call" {
			kw := elixirCallName(c, src)
			if kw != "def" && kw != "defp" && kw != "defmacro" && kw != "defmacrop" {
				return c
			}
		}
	}
	return nil
}

func elixirHeadName(head *sitter.Node, src []byte) string {
	for i := 0; i < int(head.ChildCount()); i++ {
		if head.Child(i).Type() == "identifier" {
			return nodeContent(head.Child(i), src)
		}
	}
	txt := nodeContent(head, src)
	if idx := strings.Index(txt, "("); idx > 0 {
		tok := strings.TrimSpace(txt[:idx])
		parts := strings.Fields(tok)
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return tok
	}
	return strings.TrimSpace(txt)
}

func elixirHeadParams(head *sitter.Node, src []byte) []Param {
	args := argsNode(head)
	if args == nil {
		return nil
	}
	var out []Param
	for i := 0; i < int(args.ChildCount()); i++ {
		arg := args.Child(i)
		if isPunctType(arg.Type()) {
			continue
		}
		text := strings.TrimSpace(nodeContent(arg, src))
		if text == "" {
			continue
		}
		if idx := strings.Index(text, "\\"); idx >= 0 { // keep original behavior
			name := strings.TrimSpace(text[:idx])
			def := strings.TrimSpace(text[idx+2:])
			out = append(out, Param{Name: name, Default: def})
		} else {
			out = append(out, Param{Name: text})
		}
	}
	return out
}

func extractElixirDoc(symType string, n *sitter.Node, src []byte) string {
	switch symType {
	case "module", "protocol", "impl":
		if doc := elixirModuledocInBlock(n, src); doc != "" {
			return doc
		}
		return leadingDocComments("elixir", n, src)
	case "function", "macro":
		if doc := elixirLeadingDocAttribute(n, src); doc != "" {
			return doc
		}
		return leadingDocComments("elixir", n, src)
	default:
		return ""
	}
}

func elixirModuledocInBlock(defmodule *sitter.Node, src []byte) string {
	var block *sitter.Node
	if b := defmodule.ChildByFieldName("block"); b != nil {
		block = b
	} else {
		for i := 0; i < int(defmodule.ChildCount()); i++ {
			if bt := defmodule.Child(i); bt.Type() == "do_block" || bt.Type() == "block" {
				block = bt
				break
			}
		}
	}
	if block == nil {
		return ""
	}
	for i := 0; i < int(block.ChildCount()); i++ {
		c := block.Child(i)
		txt := strings.TrimSpace(nodeContent(c, src))
		if strings.HasPrefix(txt, "@moduledoc") {
			if s := extractFirstStringLiteralFromText(txt); s != "" {
				return s
			}
			acc := txt
			for j := i + 1; j < int(block.ChildCount()) && j < i+5; j++ {
				acc += "\n" + nodeContent(block.Child(j), src)
				if s := extractFirstStringLiteralFromText(acc); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func elixirLeadingDocAttribute(n *sitter.Node, src []byte) string {
	start := n.StartByte()
	for ps := n.PrevSibling(); ps != nil; ps = ps.PrevSibling() {
		txt := strings.TrimSpace(nodeContent(ps, src))
		if txt == "" {
			continue
		}
		if hasBlankLineBetween(src, ps.EndByte(), start) {
			break
		}
		if strings.HasPrefix(txt, "@doc") {
			if strings.HasPrefix(txt, "@doc false") || strings.Contains(txt, "@doc false") {
				return ""
			}
			if s := extractFirstStringLiteralFromText(txt); s != "" {
				return s
			}
			acc := txt
			for range 5 {
				if q := ps.NextSibling(); q != nil && q != n {
					acc += "\n" + nodeContent(q, src)
					if s := extractFirstStringLiteralFromText(acc); s != "" {
						return s
					}
				} else {
					break
				}
			}
		}
		if !strings.HasPrefix(txt, "@") && !isCommentNode("elixir", ps.Type()) {
			break
		}
	}
	return ""
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

/* ---------- small helpers to reduce duplication ---------- */

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
