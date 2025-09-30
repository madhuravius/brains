init:
	go mod download
	go generate ./...
.PHONY: init

build:
	echo "Dependencies used: "
	go list -m all
	echo "Building"
	go build -v -o build/brains ./cmd/cli/*.go
.PHONY: build

tools:
	go install gotest.tools/gotestsum@latest
.PHONY: tools

lint:
	golangci-lint run
.PHONY: lint

test:
	go run gotest.tools/gotestsum@latest --format testname -- -coverprofile=cover.out ./...
.PHONY: test

pretty:
	go fmt ./...
.PHONY: pretty

release:
	goreleaser release
.PHONY: release
