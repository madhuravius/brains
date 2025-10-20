package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/google/go-github/v76/github"
	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/core"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
)

const (
	glob         = "cmd/debug/commands/ask/main.go"
	modelID      = "openai.gpt-oss-120b-1:0"
	Owner        = "madhuravius"
	IssueRefiner = `
You are a GitHub Issue Refiner.

Input:
- Original Title:
%s

- Original Body (Markdown):
%s

Goals:
- Make the title and body as clear and concise as possible while preserving the original intent and scope.
- Keep wording close to the original; do not introduce new requirements or remove essential details.
- Prefer minimal edits: compress, clarify, de-duplicate; retain technical terms, file paths, code, and links.
- If the body is empty, create a minimal body from the title without adding scope.

Tasks:
1) Title: Rewrite to a single, imperative, scope-limited line (≈5–12 words) that preserves meaning and key terms.
2) Body: Tighten wording, remove redundancy, keep structure simple. Maintain code blocks, checklists, and references.
3) Edits Made: At the very end of the body, create or update an "Edits Made" section with 1–2 sentences describing your changes. If such a section exists, append a new sentence to it.
4) Suggested Format: Provide a minimal template others can reuse to keep issues simple.

Output strictly in this structure (no extra commentary, in this exact order):

EditedTitle: <single line>

EditedBody:
<final markdown body, ending with an "Edits Made" section containing 1–2 sentences>

SuggestedSimpleFormat:
Title: Verb + object + scope (5–12 words)
Body:
- Goal: one sentence stating the desired outcome
- Scope: bullets for in-scope / out-of-scope (keep brief)
- Steps/Plan: short ordered list (optional)
- Acceptance Criteria: 2–5 concise bullet checks
- Context/Links: only essential references
`
)

func main() {
	ctx := context.Background()
	client := github.NewClient(nil)
	brainsConfig, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	awsConfig := aws.NewAWSConfig(brainsConfig.GetConfig().AWSRegion)
	awsConfig.SetLogger(brainsConfig.GetConfig())
	if !awsConfig.SetAndValidateCredentials() {
		pterm.Error.Println("unable to validate credentials")
		os.Exit(1)
	}

	coreConfig := core.NewCoreConfig(awsConfig, brainsConfig)
	coreConfig.SetLogger(brainsConfig.GetConfig())

	personaInstructions := brainsConfig.GetPersonaInstructions("dev")

	issues, _, err := client.Issues.ListByRepo(ctx, Owner, "brains", nil)
	if err != nil {
		pterm.Error.Printf("failed to get issues by repository: %v\n", err)
		os.Exit(1)
	}

	if len(issues) == 0 {
		pterm.Error.Printf("no issues found! expected issues to be printed")
		os.Exit(1)
	}

	issue := issues[0]

	pterm.Info.Printfln("issue: %s (updated: %s)", *issue.Title, issue.GetUpdatedAt())
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
		glamour.WithColorProfile(termenv.ANSI256),
	)
	result, _ := r.Render(issue.GetBody())
	pterm.Info.Println(result)

	req := &core.LLMRequest{
		Glob:                glob,
		ModelID:             modelID,
		PersonaInstructions: personaInstructions,
		Prompt:              fmt.Sprintf(IssueRefiner, issue.GetTitle(), issue.GetBody()),
	}

	if err := coreConfig.AskFlow(ctx, req); err != nil {
		pterm.Error.Printf("failed to run ask flow: %v\n", err)
		os.Exit(1)
	}

}
