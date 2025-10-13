# Brains

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/madhuravius/brains)](https://goreportcard.com/report/github.com/madhuravius/brains)

A simple CLI that wraps around AWS Bedrock and provides convenient abstractions for common LLM workflows. It is designed to help you experiment with Bedrock, explore agentic tooling, and prototype features similar to those found in tools like [Aider](https://github.com/Aider-AI/aider).

__Disclaimer__: Much like Aider, the goal of this project is for me to learn, build, and to ideally get this project to build itself (see [this page](https://aider.chat/HISTORY.html)). _This is a learning exercise and while I intend on using it for myself, use at your own risk!_

## Table of Contents
- [Prerequisites](#prerequisites)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Testing](#testing)
- [Cleaning Up](#cleaning-up)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Prerequisites
- Go 1.22+ installed.
- Valid AWS credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, optional `AWS_SESSION_TOKEN`) with permission to invoke Bedrock models.

## Features
- **Commands** to interact with Bedrock:
  - `ask` – ask questions and receive rich responses.
  - `code` – generate or modify code/files based on a natural‑language request.
- **Tools**:
  - `browser` – execute scraping functions with a Chrome‑based browser.
  - `file_system` – CRUD operations on local files.
  - `repo_map` – use tree‑sitter to produce tags, symbols, and a repository map.

## Installation
```bash
make build      # produces ./brains (or run `go build ./cmd/cli`)
```

## Usage
```bash
# Validate AWS credentials and Bedrock connectivity
./brains health

# Ask a question
./brains ask "What is AI?"

# Generate or apply code edits
./brains code "Refactor X"
```

Flags `-p/--persona` and `-a/--add` can be added to any command.

## Configuration
Create a `.brains.yml` file (the first run will generate a default one). You can set:
- `aws_region`
- `model`
- Optional personas

## Testing
```bash
make test      # runs all unit tests
```

## Cleaning Up
```bash
make clean     # removes generated binaries and logs
```

All commands are wrapped in the Makefile; run `make help` for a quick overview of available targets.

## License
This project is licensed under the MIT License – see the [LICENSE](LICENSE.md) file for details.

## Acknowledgements

Thank you to the creators of the [Aider chat tool](https://github.com/Aider-AI/aider) for the inspiration and design principles that helped shape this project.
I heavily borrowed code from [go-gitignore](https://github.com/sabhiram/go-gitignore) to avoid replicating a lot of my own.
