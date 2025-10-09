package file_system

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func (f *FileSystemConfig) UpdateFile(filePath, oldContent, newContent string, interactive bool) (bool, error) {
	if interactive {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(oldContent, newContent, false)
		diffText := dmp.DiffPrettyText(diffs)

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
			glamour.WithColorProfile(termenv.ANSI256),
		)
		renderedDiff, _ := r.Render(fmt.Sprintf("diff\n%s\n", diffText))
		fmt.Println(renderedDiff)

		ok, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Apply changes to %s?", filePath)).Show()
		if !ok {
			pterm.Warning.Printfln("Skipped: %s", filePath)
			return true, nil
		}
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		pterm.Error.Printfln("Failed to write %s: %v", filePath, err)
		return false, err
	}
	pterm.Success.Printfln("Updated %s successfully", filePath)

	return true, nil
}
