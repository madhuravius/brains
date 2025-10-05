## Default target displays help
.DEFAULT_GOAL := help

# Targets
help: ## Show this help message
	@echo "Brains Make Targets"
	@echo "Usage: make [target]"
	@echo "Common start‑up steps: \"make init\" and \"make build\""
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-30s %s\n", $$1, $$2}'
.PHONY: help

init: ## Install Go module dependencies and run generators
	go mod download
	go generate ./...
.PHONY: init


build: ## Compile the binary into ./build/brains
	echo "Dependencies used: "
	go list -m all
	echo "Building"
	go build -v -o build/brains ./cmd/cli/*.go
.PHONY: build

tools: ## Install auxiliary dev tools (gotestsum)
	go install gotest.tools/gotestsum@latest
.PHONY: tools

lint: ## Run golangci‑lint over the codebase
	golangci-lint run
.PHONY: lint

test: tools ## Execute unit tests with coverage
	go run gotest.tools/gotestsum@latest --format testname -- -coverprofile=cover.out ./...
.PHONY: test

pretty: ## Run gofmt to format source files
	go fmt ./...
.PHONY: pretty

release: ## Build and publish a new release via GoReleaser
	goreleaser release
.PHONY: release

