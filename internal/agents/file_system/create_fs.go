package file_system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func (f FileSystemConfig) CreateFile(filePath, fileContents string) error {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain("", fileContents, false)
	diffText := dmp.DiffPrettyText(diffs)

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
	fmt.Println(renderedDiff)

	// Ensure the target directory exists.
	if dir := filepath.Dir(filePath); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			pterm.Error.Printfln("Failed to create directory %s: %v", dir, err)
			return err
		}
	}

	if err := os.WriteFile(filePath, []byte(fileContents), 0644); err != nil {
		pterm.Error.Printfln("Failed to create %s: %v", filePath, err)
		return err
	}
	pterm.Success.Printfln("Created %s successfully", filePath)
	return nil
}
