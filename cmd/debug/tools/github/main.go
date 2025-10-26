package main

import (
	"context"

	"github.com/charmbracelet/glamour"
	"github.com/google/go-github/v76/github"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
)

const Owner = "madhuravius"

func main() {
	ctx := context.Background()
	client := github.NewClient(nil)

	repos, _, err := client.Repositories.ListByUser(ctx, Owner, nil)
	if err != nil {
		pterm.Fatal.Printfln("error in organization list retrieval: %v", err)
	}

	for _, repo := range repos {
		pterm.Info.Printfln("repo: %s (updated: %s)", *repo.Name, repo.GetUpdatedAt())
	}

	issues, _, err := client.Issues.ListByRepo(ctx, Owner, "brains", nil)
	if err != nil {
		pterm.Fatal.Printfln("error in issues list retrieval: %v", err)
	}
	for _, issue := range issues {
		pterm.Info.Printfln("issue: %s (updated: %s)", *issue.Title, issue.GetUpdatedAt())
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(120),
			glamour.WithColorProfile(termenv.ANSI256),
		)
		result, _ := r.Render(issue.GetBody())
		pterm.Info.Println(result)
	}
}
