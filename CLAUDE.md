# golang-sample - Go Development Guide

Production-ready Go API demonstrating clean architecture principles with Echo framework and govern package stack.

## Project Setup

**Prerequisites:**
- mise installed (manages Go version and tools)
- Go 1.25.5 (managed by mise)
- Docker (for PostgreSQL)

**Initial Setup:**
```bash
# Install mise tools
mise install

# Verify Go version
mise exec -- go version

# Install dependencies
mise exec -- go mod download
```

## Development Workflow

### Before Making Changes

1. **Install/update mise tools:**
   ```bash
   mise install
   ```

2. **Check current branch:**
   ```bash
   git status
   git branch --show-current
   ```

### Making Code Changes

1. **Edit Go files** - Follow clean architecture layers
2. **Format code:**
   ```bash
   mise exec -- goimports -w .
   ```

3. **Run static analysis:**
   ```bash
   mise exec -- golangci-lint run
   mise exec -- staticcheck ./...
   mise exec -- errcheck -blank ./...
   ```

4. **Run tests:**
   ```bash
   mise exec -- go test ./...
   mise exec -- go test -race ./...
   mise exec -- go test -cover ./...
   ```

5. **Build to verify:**
   ```bash
   mise exec -- go build ./...
   ```

### After Making Changes

1. **Update wire dependencies (if added new services):**
   ```bash
   rm internal/handler/rest/wire_gen.go
   mise exec -- wire ./internal/handler/rest/
   ```

2. **Update mocks (if added new interfaces):**
   ```bash
   mise exec -- mockery
   ```

3. **Commit with conventional commits:**
   ```bash
   git add .
   git commit -m "feat: add user registration endpoint"
   ```

## Architecture Rules

### Layer Separation

**HTTP Handlers (`internal/handler/rest/`):**
- MAY use Echo framework
- MAY bind/validate HTTP requests
- MUST NOT contain business logic
- MUST call controllers for operations

**Controllers (`internal/handler/rest/controllers/`):**
- MAY accept `echo.Context`
- MUST return schema objects (not domain models)
- MUST delegate to services
- MUST convert between schemas and service requests

**Services (`internal/service/`):**
- MUST contain business logic
- MUST be framework-agnostic
- MUST use interfaces for external dependencies
- MUST accept/return domain models

**Domain Models (`internal/model/`):**
- MUST be pure entities with NO external dependencies
- MUST NOT import GORM, Echo, or frameworks
- MAY contain business logic methods

**Storage (`internal/storage/`):**
- MUST define interfaces for data access
- MUST use GORM for database operations
- MUST return domain models (not ORM entities)

### File Naming

- **Go files:** `snake_case` (`user_service.go`)
- **Test files:** `source_test.go` (`user_service_test.go`)
- **Packages:** `snake_case` (`package user_service`)

## Testing Requirements

**Before committing:**
- All tests must pass: `mise exec -- go test ./...`
- Coverage goal: 80%+
- Race detector must pass: `mise exec -- go test -race ./...`

**Test patterns:**
- Table-driven tests for multiple cases
- Mock external dependencies
- Test both success and error paths

## Pre-commit Hooks

Automatically run on commit:
- goimports formatting
- go mod tidy check
- golangci-lint
- staticcheck
- errcheck

**Install hooks:**
```bash
mise exec -- pre-commit install
```

**Run manually:**
```bash
mise exec -- pre-commit run --all-files
```

## Common Commands

```bash
# Development
mise exec -- go run cmd/serverd.go              # Run server
mise exec -- go test ./...                       # Run tests
mise exec -- go build -o bin/serverd .           # Build binary

# Wire generation
mise exec -- wire ./internal/handler/rest/

# Mock generation
mise exec -- mockery

# Swagger docs
./scripts/generate-swagger.sh

# Database (Docker)
docker compose up -d postgres
docker compose down
```

## Dependencies

**Managed by mise (mise.toml):**
- Go 1.25.5
- wire (dependency injection)
- goimports (formatting)
- staticcheck (static analysis)
- errcheck (error checking)
- golangci-lint (linting)
- mockery (mock generation)
- swag (swagger docs)

**Adding new Go dependencies:**
```bash
mise exec -- go get github.com/package/name
mise exec -- go mod tidy
```

## Troubleshooting

**mise not found:**
```bash
curl https://mise.run | sh
# or
brew install mise
```

**Tests fail after wire changes:**
```bash
rm internal/handler/rest/wire_gen.go
mise exec -- wire ./internal/handler/rest/
mise exec -- go test ./...
```

**Mocks outdated:**
```bash
mise exec -- mockery
```

**Import issues:**
```bash
mise exec -- go mod tidy
mise exec -- go mod verify
```

## Git Workflow

**Branch naming:**
- `feat/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `chore/task-name` - Maintenance

**Commit messages:**
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring
- `test:` Test changes
- `docs:` Documentation
- `chore:` Maintenance

## Project Skills

### Active Project Skills
- **golang-dev**: Go development workflow with mise
- **go-clean-arch-template**: Template for new Go projects

### Available Golang Skills (47 skills)

**Core Development:**
- golang-code-style, golang-naming, golang-lint, golang-modernize, golang-stay-updated

**Architecture & Design:**
- golang-design-patterns, golang-dependency-injection, golang-google-wire, golang-uber-dig, golang-uber-fx, golang-samber-do, golang-project-layout

**Data Structures:**
- golang-data-structures, golang-structs-interfaces, golang-concurrency, golang-context

**Performance & Testing:**
- golang-benchmark, golang-performance, golang-testing, golang-stretchr-testify, golang-observability, golang-troubleshooting

**Error Handling & Safety:**
- golang-error-handling, golang-samber-oops, golang-samber-slog, golang-safety, golang-security

**Database & Network:**
- golang-database, golang-grpc, golang-graphql, golang-swagger

**Dependency Management:**
- golang-dependency-management, golang-popular-libraries, golang-pkg-go-dev

**CLI & Applications:**
- golang-cli, golang-spf13-cobra, golang-spf13-viper

**Documentation:**
- golang-documentation, golang-how-to

**Resilience & Utilities (Samber Series):**
- golang-samber-hot, golang-samber-lo, golang-samber-mo, golang-samber-ro

**CI/CD:**
- golang-continuous-integration

**Usage:** Skills auto-activate when working on Go files. Reference specific skills in implementation plans for domain expertise.

## Security Notes

**Never commit:**
- `.env` files with real credentials
- Generated secrets (JWT tokens, API keys)
- Temporary files or binaries

**Always:**
- Use `.env.example` with placeholders
- Generate new secrets for each environment
- Keep production secrets in secure vault
