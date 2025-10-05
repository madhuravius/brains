# Brains

A simple CLI that wraps around Bedrock and generates some abstractions to work around 
common LLMs.

## Prerequisites
- Go 1.22+ installed.
- Valid AWS credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, optional `AWS_SESSION_TOKEN`) with permission to invoke Bedrock models.

## Build

```bash
make build      # produces ./brains (or `go build ./cmd/cli`)
```


## Run
```bash
./brains health               # validates AWS creds and Bedrock connectivity
./brains ask "What is AI?"    # sends a prompt, prints response
./brains code "Refactor X"    # generates and applies code edits

Flags `-p/--persona` and `-a/--add` can be added to any command.
```

## Configuration
- Create a `.brains.yml` (first run generates a default file).
- Set `aws_region`, `model`, and optional personas in that file.

## Tests
```bash
make test      # runs all unit tests
```


## Clean
```bash
make clean     # removes generated binaries and logs
```

All commands are wrapped in the Makefile; use `make help` for a quick overview.
