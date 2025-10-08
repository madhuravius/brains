package main

import (
	"context"

	"brains/internal/tools/browser"
	"brains/internal/tools/file_system"

	"github.com/pterm/pterm"
)

func main() {
	fs, err := file_system.NewFileSystemConfig()
	if err != nil {
		pterm.Fatal.Printfln("file_system.NewFileSystemConfig: %v", err)
	}

	fsData, err := fs.SetContextFromGlob("README.md")
	if err != nil {
		pterm.Fatal.Printfln("file_system.SetContextFromGlob: %v", err)
	}
	pterm.Info.Printfln("data from glob gather: %s", fsData)

	htmlData, err := browser.FetchWebContext(context.Background(), "https://github.com/madhuravius")
	if err != nil {
		pterm.Fatal.Printfln("browser.FetchWebContext: %v", err)
	}
	pterm.Info.Printfln("data from web gather: %s", htmlData)
}
