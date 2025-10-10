package repo_map_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madhuravius/brains/internal/tools/repo_map"
)

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected to contain %q\nGot:\n%s", sub, s)
	}
}

func shouldContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected to contain %q\nGot:\n%s", sub, s)
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

func TestGo_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "go", "go/sample.go")

	// Symbols
	shouldContain(t, out, "- struct Person")
	shouldContain(t, out, "- interface Greeter")
	shouldContain(t, out, "- constant Pi")
	shouldContain(t, out, "- variable version")
	shouldContain(t, out, "- function SayHello")
	shouldContain(t, out, "- method Greet")

	// Params (coarse)
	shouldContain(t, out, "SayHello(")
	shouldContain(t, out, "name")
	shouldContain(t, out, "times")

	// Docs (from leading comments)
	shouldContain(t, out, "Doc: SayHello says hi")
}

func TestPython_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "python", "python/sample.py")

	// Symbols
	shouldContain(t, out, "- class Person")
	shouldContain(t, out, "- function greet")

	// Params
	shouldContain(t, out, "greet(name")

	// Docs (docstrings)
	shouldContain(t, out, "Doc: Greet someone.")
	shouldContain(t, out, "Doc: Represents a person.")
}

func TestJavaScript_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "javascript", "javascript/sample.js")

	// Symbols
	shouldContain(t, out, "- function greet")
	shouldContain(t, out, "- class Person")
	shouldContain(t, out, "- method speak")

	// Params
	shouldContain(t, out, "greet(name")
	shouldContain(t, out, "speak(msg")

	// Docs (JSDoc)
	shouldContain(t, out, "Doc: Say hi")
	shouldContain(t, out, "Doc: Speak something")
}

func TestTypeScript_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "typescript", "typescript/sample.ts")

	// Symbols
	shouldContain(t, out, "- function doThing")
	shouldContain(t, out, "- class MyClass")
	shouldContain(t, out, "- interface IThing")

	// Params (typed + rest). Use partial contains for stability.
	shouldContain(t, out, "doThing(")
	shouldContain(t, out, "x: number")
	shouldContain(t, out, "...rest")

	// Docs
	shouldContain(t, out, "Doc: Do a thing")
}

func TestJava_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "java", "java/sample.java")

	// Symbols
	shouldContain(t, out, "- class Service")
	shouldContain(t, out, "- constructor Service")
	shouldContain(t, out, "- method execute")

	// Params (typed + varargs). Use partials for stability.
	shouldContain(t, out, "execute(")
	shouldContain(t, out, "value: String")
	// If your formatting is "String value", keep this alternative:
	// shouldContain(t, out, "String")
	// shouldContain(t, out, "value")
	shouldContain(t, out, "...") // spread/varargs marker appears on one param

	// Docs (Javadoc)
	shouldContain(t, out, "Doc: Service class")
	shouldContain(t, out, "Doc: Execute work")
}

func TestCPP_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "cpp", "cpp/sample.cpp")

	// Symbols
	shouldContain(t, out, "- function add")
	shouldContain(t, out, "- class Box")
	shouldContain(t, out, "- struct Data")

	// Params (typed)
	shouldContain(t, out, "add(")
	shouldContain(t, out, "a: int")
	shouldContain(t, out, "b: int")
}

func TestCSharp_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "csharp", "csharp/sample.cs")

	// Symbols
	shouldContain(t, out, "- interface IRepo")
	shouldContain(t, out, "- struct Point")
	shouldContain(t, out, "- class Person")
	shouldContain(t, out, "- method Speak")

	// Params (modifiers: this, ref, out, params)
	shouldContain(t, out, "Speak(")
	shouldContain(t, out, "this ")
	shouldContain(t, out, "ref ")
	shouldContain(t, out, "out ")
	shouldContain(t, out, "params ")

	// Docs (triple-slash)
	shouldContain(t, out, "Doc: Speak doc")
}

func TestRuby_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "ruby", "ruby/sample.rb")

	// Symbols
	shouldContain(t, out, "- class Person")
	shouldContain(t, out, "- method speak")
	shouldContain(t, out, "- module MyMod")

	// Params: required, splat, keyword splat, block
	shouldContain(t, out, "speak(")
	shouldContain(t, out, "name")
	shouldContain(t, out, "*rest")
	shouldContain(t, out, "**kw")
	shouldContain(t, out, "&blk")

	// Docs (leading # comments)
	shouldContain(t, out, "Doc: Person doc")
	shouldContain(t, out, "Doc: speak doc")
}

func TestPHP_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "php", "php/sample.php")

	// Symbols
	shouldContain(t, out, "- function greet")
	shouldContain(t, out, "- class Person")
	shouldContain(t, out, "- method speak")
	shouldContain(t, out, "- interface IThing")

	// Params: typed, by-ref, variadic
	shouldContain(t, out, "greet(")
	shouldContain(t, out, "name: string")
	shouldContain(t, out, "&$ref")
	shouldContain(t, out, "...$rest")

	// Docs (PHPDoc)
	shouldContain(t, out, "Doc: Greet doc")
	shouldContain(t, out, "Doc: speak doc")
}

func TestRust_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "rust", "rust/sample.rs")

	// Symbols
	shouldContain(t, out, "- struct Person")
	shouldContain(t, out, "- function add")
	shouldContain(t, out, "- trait Speak")

	// Params (typed)
	shouldContain(t, out, "add(")
	shouldContain(t, out, "x: i32")
	shouldContain(t, out, "y: i32")

	// Docs (///)
	shouldContain(t, out, "Doc: Person doc")
	shouldContain(t, out, "Doc: add doc")
	shouldContain(t, out, "Doc: Speak trait")
}

func TestElixir_DocsAndParams(t *testing.T) {
	t.Parallel()
	out := renderSingleFile(t, "elixir", "elixir/sample.ex")

	// Symbols
	shouldContain(t, out, "- module Greeter")
	shouldContain(t, out, "- function hello")
	shouldContain(t, out, "- function private_thing")
	shouldContain(t, out, "- macro debug")

	// Params (function head). Defaults parsed but not necessarily rendered.
	shouldContain(t, out, "hello(")
	shouldContain(t, out, "name")
	shouldContain(t, out, "private_thing(")
	shouldContain(t, out, "x")
	shouldContain(t, out, "debug(")
	shouldContain(t, out, "expr")

	// Docs (@moduledoc and @doc)
	shouldContain(t, out, "Doc: Greeter module docs.")
	shouldContain(t, out, "Doc: Says hello.")
	shouldContain(t, out, "Doc: Debug macro.")
}
