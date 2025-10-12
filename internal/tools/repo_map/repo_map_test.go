package repo_map_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madhuravius/brains/internal/tools/repo_map"
	"github.com/stretchr/testify/assert"
)

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected to contain %q\nGot:\n%s", sub, s)
	}
}

func renderSingleFile(t *testing.T, lang, relpath string) string {
	t.Helper()
	ctx := context.Background()
	root := "test_fixtures"
	path := filepath.Join(root, relpath)

	fm, err := repo_map.ParseFile(ctx, path, lang)
	if err != nil {
		t.Fatalf("ParseFile(%q, %q) error: %v", path, lang, err)
	}
	r := repo_map.RepoMap{
		Path:  filepath.Dir(path),
		Files: []*repo_map.FileMap{fm},
	}
	out := r.ToPrompt()
	mustContain(t, out, fmt.Sprintf("### File: %s", path))
	return out
}

func TestRepoMap(t *testing.T) {
	ctx := context.Background()
	path := "test_fixtures"
	repoMap, err := repo_map.BuildRepoMap(ctx, path)

	assert.Nil(t, err)
	assert.NotNil(t, repoMap)
}

func TestGo_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "go", "go/sample.go")

	// Symbols
	assert.Contains(t, out, "- struct Person")
	assert.Contains(t, out, "- interface Greeter")
	assert.Contains(t, out, "- constant Pi")
	assert.Contains(t, out, "- variable version")
	assert.Contains(t, out, "- function SayHello")
	assert.Contains(t, out, "- method Greet")

	// Params (coarse)
	assert.Contains(t, out, "SayHello(")
	assert.Contains(t, out, "name")
	assert.Contains(t, out, "times")

	// Docs (from leading comments)
	assert.Contains(t, out, "Doc: SayHello says hi")
}

func TestPython_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "python", "python/sample.py")

	// Symbols
	assert.Contains(t, out, "- class Person")
	assert.Contains(t, out, "- function greet")

	// Params
	assert.Contains(t, out, "greet(name")

	// Docs (docstrings)
	assert.Contains(t, out, "Doc: Greet someone.")
	assert.Contains(t, out, "Doc: Represents a person.")
}

func TestJavaScript_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "javascript", "javascript/sample.js")

	// Symbols
	assert.Contains(t, out, "- function greet")
	assert.Contains(t, out, "- class Person")
	assert.Contains(t, out, "- method speak")

	// Params
	assert.Contains(t, out, "greet(name")
	assert.Contains(t, out, "speak(msg")

	// Docs (JSDoc)
	assert.Contains(t, out, "Doc: Say hi")
	assert.Contains(t, out, "Doc: Speak something")
}

func TestTypeScript_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "typescript", "typescript/sample.ts")

	// Symbols
	assert.Contains(t, out, "- function doThing")
	assert.Contains(t, out, "- class MyClass")
	assert.Contains(t, out, "- interface IThing")

	// Params (typed + rest). Use partial contains for stability.
	assert.Contains(t, out, "doThing(")
	assert.Contains(t, out, "x: number")
	assert.Contains(t, out, "...rest")

	// Docs
	assert.Contains(t, out, "Doc: Do a thing")
}

func TestJava_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "java", "java/sample.java")

	// Symbols
	assert.Contains(t, out, "- class Service")
	assert.Contains(t, out, "- constructor Service")
	assert.Contains(t, out, "- method execute")

	// Params (typed + varargs). Use partials for stability.
	assert.Contains(t, out, "execute(")
	assert.Contains(t, out, "value: String")
	// If your formatting is "String value", keep this alternative:
	// assert.Contains(t, out, "String")
	// assert.Contains(t, out, "value")
	assert.Contains(t, out, "...") // spread/varargs marker appears on one param

	// Docs (Javadoc)
	assert.Contains(t, out, "Doc: Service class")
	assert.Contains(t, out, "Doc: Execute work")
}

func TestCPP_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "cpp", "cpp/sample.cpp")

	// Symbols
	assert.Contains(t, out, "- function add")
	assert.Contains(t, out, "- class Box")
	assert.Contains(t, out, "- struct Data")

	// Params (typed)
	assert.Contains(t, out, "add(")
	assert.Contains(t, out, "a: int")
	assert.Contains(t, out, "b: int")
}

func TestCSharp_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "csharp", "csharp/sample.cs")

	// Symbols
	assert.Contains(t, out, "- interface IRepo")
	assert.Contains(t, out, "- struct Point")
	assert.Contains(t, out, "- class Person")
	assert.Contains(t, out, "- method Speak")

	// Params (modifiers: this, ref, out, params)
	assert.Contains(t, out, "Speak(")
	assert.Contains(t, out, "this ")
	assert.Contains(t, out, "ref ")
	assert.Contains(t, out, "out ")
	assert.Contains(t, out, "params ")

	// Docs (triple-slash)
	assert.Contains(t, out, "Doc: Speak doc")
}

func TestRuby_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "ruby", "ruby/sample.rb")

	// Symbols
	assert.Contains(t, out, "- class Person")
	assert.Contains(t, out, "- method speak")
	assert.Contains(t, out, "- module MyMod")

	// Params: required, splat, keyword splat, block
	assert.Contains(t, out, "speak(")
	assert.Contains(t, out, "name")
	assert.Contains(t, out, "*rest")
	assert.Contains(t, out, "**kw")
	assert.Contains(t, out, "&blk")

	// Docs (leading # comments)
	assert.Contains(t, out, "Doc: Person doc")
	assert.Contains(t, out, "Doc: speak doc")
}

func TestPHP_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "php", "php/sample.php")

	// Symbols
	assert.Contains(t, out, "- function greet")
	assert.Contains(t, out, "- class Person")
	assert.Contains(t, out, "- method speak")
	assert.Contains(t, out, "- interface IThing")

	// Params: typed, by-ref, variadic
	assert.Contains(t, out, "greet(")
	assert.Contains(t, out, "name: string")
	assert.Contains(t, out, "&$ref")
	assert.Contains(t, out, "...$rest")

	// Docs (PHPDoc)
	assert.Contains(t, out, "Doc: Greet doc")
	assert.Contains(t, out, "Doc: speak doc")
}

func TestRust_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "rust", "rust/sample.rs")

	// Symbols
	assert.Contains(t, out, "- struct Person")
	assert.Contains(t, out, "- function add")
	assert.Contains(t, out, "- trait Speak")

	// Params (typed)
	assert.Contains(t, out, "add(")
	assert.Contains(t, out, "x: i32")
	assert.Contains(t, out, "y: i32")

	// Docs (///)
	assert.Contains(t, out, "Doc: Person doc")
	assert.Contains(t, out, "Doc: add doc")
	assert.Contains(t, out, "Doc: Speak trait")
}

// repo_map_test.go

func TestElixir_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "elixir", "elixir/sample.ex")

	assert.Contains(t, out, "- module Greeter")
	assert.Contains(t, out, "Doc: Greeter module docs.")

	assert.Contains(t, out, "- function hello(name)") // Note: Renderer might not show default
	assert.Contains(t, out, "Doc: Says hello.")

	assert.Contains(t, out, "- function private_thing(x)")
	assert.Contains(t, out, "- macro debug(expr)")
	assert.Contains(t, out, "Doc: Debug macro.")

	assert.Contains(t, out, "- macro private_macro(val)")
	assert.Contains(t, out, "Doc: A private macro.")

	assert.Contains(t, out, "- protocol Parser")
	assert.Contains(t, out, "Doc: A simple protocol.")
	assert.Contains(t, out, "- function parse(data)")
	assert.Contains(t, out, "Doc: Parses the data.")

	assert.Contains(t, out, "- impl Parser") // The symbol name is the protocol
	assert.Contains(t, out, "- function parse(_data)")

	assert.Contains(t, out, "- function no_args")
	assert.Contains(t, out, "- function no_parens")

	assert.Contains(t, out, "- function undocumented_fun")
	assert.NotContains(t, out, "undocumented_fun()\n    Doc:")

	assert.Contains(t, out, "- function commented_fun")

	assert.Contains(t, out, "- function fun_with_separated_doc")
	assert.NotContains(t, out, "function fun_with_separated_doc()\n    Doc: This doc should be ignored.")
}
