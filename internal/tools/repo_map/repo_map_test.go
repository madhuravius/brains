package repo_map_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madhuravius/brains/internal/tools/repo_map"
)

func TestParseFile_ByLanguage(t *testing.T) {
	ctx := context.Background()
	root := "test_fixtures"

	cases := []struct {
		name     string
		lang     string
		relpath  string
		expected []string // substrings expected to appear in RepoMap.ToPrompt() output
	}{
		{
			name:    "go",
			lang:    "go",
			relpath: "go/sample.go",
			expected: []string{
				"struct Person",
				"interface Greeter",
				"function SayHello",
				"method Greet",
				"constant Pi",
				"variable version",
			},
		},
		{
			name:    "python",
			lang:    "python",
			relpath: "python/sample.py",
			expected: []string{
				"class Person",
				"function greet",
				"constant CONSTANT",
			},
		},
		{
			name:    "javascript",
			lang:    "javascript",
			relpath: "javascript/sample.js",
			expected: []string{
				"function greet",
				"class Person",
				"method speak",
			},
		},
		{
			name:    "typescript",
			lang:    "typescript",
			relpath: "typescript/sample.ts",
			expected: []string{
				"function doThing",
				"class MyClass",
				"interface IThing",
			},
		},
		{
			name:    "java",
			lang:    "java",
			relpath: "java/sample.java",
			expected: []string{
				"class Service",
				"method execute",
				"constructor Service",
			},
		},
		{
			name:    "cpp",
			lang:    "cpp",
			relpath: "cpp/sample.cpp",
			expected: []string{
				"function add",
				"class Box",
				"struct Data",
			},
		},
		{
			name:    "csharp",
			lang:    "csharp",
			relpath: "csharp/sample.cs",
			expected: []string{
				"class Person",
				"method Speak",
				"struct Point",
				"interface IRepo",
			},
		},
		{
			name:    "ruby",
			lang:    "ruby",
			relpath: "ruby/sample.rb",
			expected: []string{
				"class Person",
				"method speak",
				"module MyMod",
			},
		},
		{
			name:    "php",
			lang:    "php",
			relpath: "php/sample.php",
			expected: []string{
				"function greet",
				"class Person",
				"method speak",
				"interface IThing",
			},
		},
		{
			name:    "rust",
			lang:    "rust",
			relpath: "rust/sample.rs",
			expected: []string{
				"struct Person",
				"function add",
				"trait Speak",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(root, tc.relpath)

			fileMap, err := repo_map.ParseFile(ctx, path, tc.lang)
			if err != nil {
				t.Fatalf("ParseFile(%q, %q) error: %v", path, tc.lang, err)
			}

			// Build a one-file RepoMap and render it to a prompt to assert the textual output.
			repo := repo_map.RepoMap{
				Path:  filepath.Dir(path),
				Files: []*repo_map.FileMap{fileMap},
			}

			out := repo.ToPrompt()

			// check header exists (file path should appear)
			if !strings.Contains(out, fmt.Sprintf("### File: %s", path)) {
				t.Fatalf("expected ToPrompt() to include file header for %s; got:\n%s", path, out)
			}

			// each expected snippet should show up
			for _, want := range tc.expected {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q to appear in prompt for %s; output:\n%s", want, tc.relpath, out)
				}
			}
		})
	}
}
