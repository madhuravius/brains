package main

import (
	"context"

	"github.com/madhuravius/brains/internal/tools/repo_map"

	"github.com/pterm/pterm"
)

func main() {
	ctx := context.Background()
	repoMap, err := repo_map.BuildRepoMap(ctx, "./")
	if err != nil {
		pterm.Fatal.Printfln("repo_map.BuildRepoMap: %v", err)
	}
	pterm.Info.Printfln("data from repo map: %s", repoMap.ToPrompt())
}
