---
name: golang-dev
description: Go development workflow for this project with mise toolchain. Use when developing Go code in this project, running tests, building binaries, or managing dependencies. Activates on: Go file edits, go.mod changes, test commands, or any development workflow.
license: MIT
argument-hint: "[command]"
metadata:
  project: golang-sample
  go_version: "1.25.5"
  mise_version: "2024.6.0"
---

# Go Development for golang-sample

Project-specific Go development workflow with mise integration.

## Quick Start

```bash
# Install mise tools
mise install

# Run tests
mise exec -- go test ./...

# Build project
mise exec -- go build ./...

# Format code
mise exec -- goimports -w .
```

## Mise Tools

**Available via mise:**
- `go`: 1.25.5
- `wire`: Dependency injection
- `goimports`: Import management
- `staticcheck`: Static analysis
- `errcheck`: Error checking
- `golangci-lint`: Linting
- `mockery`: Mock generation
- `swag`: Swagger docs

## Common Commands

```bash
# Development workflow
mise exec -- go mod tidy
mise exec -- goimports -w .
mise exec -- go test ./...
mise exec -- go build ./...

# Wire generation
mise exec -- wire ./internal/handler/rest/

# Mock generation
mise exec -- mockery

# Swagger docs
./scripts/generate-swagger.sh

# Pre-commit
mise exec -- pre-commit run --all-files
```

## Project Structure

```
golang-sample/
├── cmd/                    # Application entry points
├── internal/               # Private code
│   ├── handler/rest/       # HTTP handlers (Echo)
│   ├── service/            # Business logic
│   ├── storage/            # Data access
│   ├── model/              # Domain models
│   └── orm/                # ORM entities
├── pkg/                    # Public libraries
├── docs/                   # Documentation
└── .github/                # GitHub workflows
```

## Testing

```bash
# All tests
mise exec -- go test ./...

# With coverage
mise exec -- go test -cover ./...

# Specific package
mise exec -- go test ./internal/service/...

# Race detection
mise exec -- go test -race ./...
```

## Building

```bash
# Build binary
mise exec -- go build -o bin/serverd .

# Run server
./bin/serverd
```

## Pre-commit

Hooks automatically run:
- goimports formatting
- go mod tidy check
- golangci-lint
- staticcheck
- errcheck

## Troubleshooting

**Tools not found:**
```bash
mise install
```

**Tests fail:**
```bash
mise exec -- go mod tidy
mise exec -- go test ./...
```

**Wire issues:**
```bash
rm internal/handler/rest/wire_gen.go
mise exec -- wire ./internal/handler/rest/
```
