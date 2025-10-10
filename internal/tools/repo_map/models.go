package repo_map

type RepoMap struct {
	Path  string     `json:"path"`
	Files []*FileMap `json:"files"`
}

type FileMap struct {
	Path     string       `json:"path"`
	Language string       `json:"language"`
	Summary  string       `json:"summary,omitempty"`
	Symbols  []*SymbolMap `json:"symbols"`
}

type SymbolMap struct {
	Type     string       `json:"type"` // e.g., "function", "class", "method", "variable"
	Name     string       `json:"name"`
	Start    int          `json:"start"`
	End      int          `json:"end"`
	Doc      string       `json:"doc,omitempty"`
	Children []*SymbolMap `json:"children,omitempty"`
}

type SymbolRule struct {
	NodeType   string
	FieldName  string
	SymbolType string
}

var LanguageSymbolRules = map[string][]SymbolRule{
	"go": {
		{"function_declaration", "name", "function"},
		{"method_declaration", "name", "method"},
		{"const_spec", "name", "constant"},
		{"var_spec", "name", "variable"},
	},

	"python": {
		{"function_definition", "name", "function"},
		{"class_definition", "name", "class"},
	},

	"javascript": {
		{"function_declaration", "name", "function"},
		{"method_definition", "name", "method"},
		{"class_declaration", "name", "class"},
		{"arrow_function", "", "function"},
	},

	"typescript": {
		{"function_declaration", "name", "function"},
		{"method_signature", "name", "method"},
		{"class_declaration", "name", "class"},
		{"interface_declaration", "name", "interface"},
	},

	"java": {
		{"method_declaration", "name", "method"},
		{"constructor_declaration", "name", "constructor"},
		{"class_declaration", "name", "class"},
		{"interface_declaration", "name", "interface"},
		{"enum_declaration", "name", "enum"},
	},

	"cpp": {
		{"function_definition", "declarator", "function"},
		{"class_specifier", "name", "class"},
		{"struct_specifier", "name", "struct"},
		{"namespace_definition", "name", "namespace"},
	},

	"csharp": {
		{"method_declaration", "name", "method"},
		{"constructor_declaration", "name", "constructor"},
		{"class_declaration", "name", "class"},
		{"struct_declaration", "name", "struct"},
		{"interface_declaration", "name", "interface"},
	},

	"ruby": {
		{"method", "name", "method"},
		{"class", "name", "class"},
		{"module", "name", "module"},
	},

	"php": {
		{"function_definition", "name", "function"},
		{"method_declaration", "name", "method"},
		{"class_declaration", "name", "class"},
		{"interface_declaration", "name", "interface"},
	},

	"rust": {
		{"function_item", "name", "function"},
		{"impl_item", "name", "impl"},
		{"struct_item", "name", "struct"},
		{"enum_item", "name", "enum"},
		{"trait_item", "name", "trait"},
	},

	"elixir": {
		{"defmodule", "name", "module"},
		{"def", "name", "function"},
		{"defp", "name", "function"},
		{"defmacro", "name", "macro"},
	},
}
