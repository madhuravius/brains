package file_system

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const FileTreeRootLabel = ". (root)"

func (f *FileSystemConfig) GetFileTree(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	if st, err := os.Stat(absRoot); err != nil || !st.IsDir() {
		return "", fmt.Errorf("root is not a directory: %s", root)
	}

	var b strings.Builder
	b.WriteString(FileTreeRootLabel)
	b.WriteByte('\n')

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if f.commonTools.IsIgnored(rel) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		for range strings.Count(rel, "/") + 1 {
			b.WriteString("  ")
		}

		name := d.Name()
		if d.IsDir() {
			b.WriteString(name)
			b.WriteString("/\n")
			return nil
		}

		b.WriteString(name)
		b.WriteByte('\n')
		return nil
	})
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
