## Default target displays help
.DEFAULT_GOAL := help

# Targets
help: ## Show this help message
	@echo "Brains Make Targets"
	@echo "Usage: make [target]"
	@echo "Common start‑up steps: \"make init\" and \"make build\""
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-30s %s\n", $$1, $$2}'
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
	go install github.com/vladopajic/go-test-coverage/v2@latest
	go install gotest.tools/gotestsum@latest
.PHONY: tools

lint: tools ## Run golangci‑lint over the codebase
	golangci-lint run
.PHONY: lint

test: tools ## Execute unit tests with coverage
	go run gotest.tools/gotestsum@latest --format testname -- -coverprofile=cover.out ./...
	go run github.com/vladopajic/go-test-coverage/v2@latest --config=./.testcoverage.yml
.PHONY: test

pretty: ## Run gofmt to format source files
	go fmt ./...
.PHONY: pretty

release: ## Build and publish a new release via GoReleaser
	goreleaser release
.PHONY: release

require-root: ## Ensure we have root privileges for actions that touch system paths
	@if [ "$$$(id -u)" -ne 0 ]; then \
		echo "Error: this target requires root. Run with sudo."; \
		exit 1; \
	fi
.PHONY: require-root

install: require-root ## Build and install the binary to /usr/local/bin
	@$(MAKE) build
	sudo install -m 0755 ./build/brains $(DESTDIR)/usr/local/bin/brains
.PHONY: install

