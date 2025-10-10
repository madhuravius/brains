package main

import (
	"context"

	"github.com/madhuravius/brains/internal/tools/browser"
	"github.com/madhuravius/brains/internal/tools/file_system"
	"github.com/madhuravius/brains/internal/tools/repo_map"

	"github.com/pterm/pterm"
)

func main() {
	ctx := context.Background()
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

	repoMap, err := repo_map.BuildRepoMap(ctx, "./")
	if err != nil {
		pterm.Fatal.Printfln("repo_map.BuildRepoMap: %v", err)
	}
	pterm.Info.Printfln("data from repo map: %s", repoMap.ToPrompt())
}
