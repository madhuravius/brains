package main

import (
	"context"

	"github.com/madhuravius/brains/internal/tools/browser"

	"github.com/pterm/pterm"
)

func main() {
	b, err := browser.NewBrowserConfig()
	if err != nil {
		pterm.Fatal.Printfln("browser.NewBrowserConfig: %v", err)
	}
	htmlData, err := b.FetchWebContext(context.Background(), "https://github.com/madhuravius")
	if err != nil {
		pterm.Fatal.Printfln("browser.FetchWebContext: %v", err)
	}
	pterm.Info.Printfln("data from web gather: %s", htmlData)
}
