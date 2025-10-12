package repo_map

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

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
		} else {
			name = elixirFirstArgText(n, src)
		}
	case "defmacro", "defmacrop":
		symType = "macro"
		if head := elixirFindHeadCall(n, src); head != nil {
			name = elixirHeadName(head, src)
			params = elixirHeadParams(head, src)
		} else {
			name = elixirFirstArgText(n, src)
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
	return ""
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
	case "module", "impl":
		if doc := elixirModuledocInBlock(n, src); doc != "" {
			return doc
		}
		return leadingDocComments("elixir", n, src)
	case "function", "macro", "protocol":
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
