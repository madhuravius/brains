package repo_map

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func cppParams(n *sitter.Node, src []byte) []Param {
	plist := childByFieldOrTypes(n, "parameters", "parameter_list")
	if plist == nil {
		plist = findFirstByType(n, "parameter_list")
	}
	if plist == nil {
		return nil
	}
	size := int(plist.ChildCount())
	out := make([]Param, size)
	for i := range size {
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

func jsLikeFallbackParams(text string) []Param {
	start := strings.Index(text, "(")
	end := strings.LastIndex(text, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return nil
	}
	inner := text[start+1 : end]
	parts := strings.Split(inner, ",")
	out := make([]Param, len(parts))
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
