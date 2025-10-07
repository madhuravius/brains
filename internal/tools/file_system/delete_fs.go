package file_system

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func (f *FileSystemConfig) DeleteFile(filePath string) error {
	oldContentBytes, readErr := os.ReadFile(filePath)
	if readErr != nil {
		pterm.Error.Printfln("Failed to read %s for diff: %v", filePath, readErr)
	} else {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(oldContentBytes), "", false)
		diffText := dmp.DiffPrettyText(diffs)

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
			glamour.WithColorProfile(termenv.ANSI256),
		)
		renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
		fmt.Println(renderedDiff)
	}

	if err := os.Remove(filePath); err != nil {
		pterm.Error.Printfln("Failed to delete %s: %v", filePath, err)
		return err
	}
	pterm.Success.Printfln("Deleted %s successfully", filePath)
	return nil
}
