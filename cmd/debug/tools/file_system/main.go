package main

import (
	"github.com/madhuravius/brains/internal/tools/file_system"

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

	treeData, err := fs.GetFileTree("./")
	if err != nil {
		pterm.Fatal.Printfln("file_system.GetFileTree: %v", err)
	}
	pterm.Info.Printfln("data from file tree: \n\n%s", treeData)

}
