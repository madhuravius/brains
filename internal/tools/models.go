package tools

type CommonToolsImpl interface {
	IsIgnored(path string) bool
}

type CommonToolsConfig struct {
	ignorePatterns []string
}

var defaultIgnoreNames = map[string]struct{}{
	"package-lock.json": {},
	"yarn.lock":         {},
	"pnpm-lock.yaml":    {},
	"bun.lockb":         {},
	"go.sum":            {},
	"poetry.lock":       {},
	"Cargo.lock":        {},
	"Gemfile.lock":      {},
	"composer.lock":     {},
	"Pipfile.lock":      {},
	"mix.lock":          {},
	"Podfile.lock":      {},
	"package.json.lock": {},
	"flake.lock":        {},
	"requirements.txt":  {},
	"target":            {},
	"node_modules":      {},
	".venv":             {},
}
