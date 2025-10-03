package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/pterm/pterm"
)

func (b *BrainsConfig) GetPersonaInstructions(persona string) string {
	personaText, found := b.Personas[persona]
	if !found {
		return ""
	}
	pterm.Info.Printfln("user electing to leverage persona (%s) with text: %s", persona, personaText)
	return fmt.Sprintf("Human: %s\n\n", personaText)
}

func (b *BrainsConfig) SetContextFromGlob(pattern string) (string, error) {
	files, err := doublestar.Glob(os.DirFS("."), pattern)
	if err != nil {
		return "", fmt.Errorf("failed to expand glob: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files matched pattern %s", pattern)
	}

	contents := make([]string, 0, len(files)*2)

	for idx, fpath := range files {
		info, err := os.Stat(fpath)
		if err != nil {
			pterm.Warning.Printfln("failed to stat %s: %v", fpath, err)
			continue
		}
		if info.IsDir() {
			continue
		}

		data, err := os.ReadFile(fpath)
		if err != nil {
			pterm.Warning.Printfln("failed to read %s: %v", fpath, err)
			continue
		}

		if idx == 0 {
			contents = append(contents, "\n\n")
		}
		pterm.Info.Printfln("added file to context: %s", fpath)
		contents = append(contents,
			fmt.Sprintf("\n\n--- %s ---\n%s", filepath.Base(fpath), string(data)))
	}

	return strings.Join(contents, ""), nil
}
