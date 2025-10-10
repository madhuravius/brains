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
		{"type_declaration", "name", "type"},
	},

	"python": {
		{"function_definition", "name", "function"},
		{"class_definition", "name", "class"},
	},

	"javascript": {
		{"function_declaration", "name", "function"},
		{"class_declaration", "name", "class"},
	},

	"typescript": {
		{"function_declaration", "name", "function"},
		{"class_declaration", "name", "class"},
	},

	"java": {
		{"method_declaration", "name", "method"},
		{"class_declaration", "name", "class"},
		{"interface_declaration", "name", "interface"},
	},

	"cpp": {
		{"function_definition", "declarator", "function"},
		{"class_specifier", "name", "class"},
		{"struct_specifier", "name", "struct"},
	},

	"csharp": {
		{"method_declaration", "name", "method"},
		{"class_declaration", "name", "class"},
		{"struct_declaration", "name", "struct"},
	},

	"ruby": {
		{"method", "name", "method"},
		{"class", "name", "class"},
		{"module", "name", "module"},
	},

	"php": {
		{"function_definition", "name", "function"},
		{"class_declaration", "name", "class"},
	},

	"rust": {
		{"function_item", "name", "function"},
		{"impl_item", "name", "impl"},
		{"struct_item", "name", "struct"},
		{"enum_item", "name", "enum"},
	},

	"elixir": {
		{"module", "name", "module"},
		{"function_clause", "name", "function"},
		{"function_definition", "name", "function"},
		{"def", "name", "function"},
		{"defp", "name", "function"},
		{"defmodule", "name", "module"},
	},
}
