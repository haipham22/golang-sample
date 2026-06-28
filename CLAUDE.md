# Govern - Go Development Guide

Monorepo containing the **govern** library (root module) and a sample application
(`examples/golang-sample/`). Both are independent Go modules — there is **no `go.work`**; the
sample imports govern as an external dependency via a `replace` directive.

## Repository Structure

```
govern/                              # Module: github.com/haipham22/govern
├── http/  database/  config/  errors/  log/  graceful/
├── retry/  cron/  mq/  metrics/  healthcheck/
├── go.mod                           # module github.com/haipham22/govern
├── Makefile                         # govern library targets (test/build/lint)
├── README.md  CLAUDE.md  CONTRIBUTING.md
├── docs/                            # govern library docs
├── .claude/rules/                   # Go development rules
└── examples/
    └── golang-sample/               # Module: github.com/haipham22/golang-sample
        ├── cmd/  internal/  pkg/
        ├── go.mod                   # require govern v0.0.0 + replace => ../../
        └── Makefile                 # sample app targets
```

## Modules

- **Govern Library** (root) — `github.com/haipham22/govern`, published as a Go library.
- **Sample App** (`examples/golang-sample/`) — `github.com/haipham22/golang-sample`, demonstrates
  govern usage with clean architecture. Uses `replace github.com/haipham22/govern => ../../` for
  local development.

## Prerequisites

- [mise](https://mise.run) (manages Go version and tools)
- Go 1.25+ (managed by mise)
- Docker (for PostgreSQL when running the sample app)

## Initial Setup

```bash
mise install                    # Install Go + tools
mise exec -- go version         # Verify Go version
mise exec -- go mod download    # Fetch govern library deps
```

## Development Workflow

### Govern Library (root)

```bash
mise exec -- goimports -w .                 # Format
mise exec -- golangci-lint run              # Lint
mise exec -- staticcheck ./...              # Static analysis
mise exec -- errcheck -blank ./...          # Error check
mise exec -- go test ./...                  # Test
mise exec -- go test -race ./...            # Test (race)
mise exec -- go build ./...                 # Build
```

Or via Makefile: `make test | build | lint`.

### Sample App (`examples/golang-sample/`)

```bash
cd examples/golang-sample
mise exec -- go mod tidy                     # Resolve deps (incl. local govern via replace)
mise exec -- go test ./...                   # Test
mise exec -- go build -o bin/serverd .       # Build
mise exec -- mockery                         # Regenerate mocks (if interfaces changed)
```

See [`examples/golang-sample/CLAUDE.md`](examples/golang-sample/CLAUDE.md) for sample-app specifics.

## Development Rules

Comprehensive Go rules live in [`.claude/rules/`](.claude/rules/) — see
[`.claude/rules/README.md`](.claude/rules/README.md) for the full overview:

- Type system (pointers, generics, values)
- Context & concurrency
- Error handling & wrapping
- Validation (go-playground/validator)
- Database & GORM
- Swagger/OpenAPI
- Testing
- Clean architecture
- Docker, mise toolchain, infrastructure

**Quick reference — key rules:**
- ✅ Always use `mise exec --` for Go commands
- ✅ Pass `context.Context` as first parameter in I/O functions
- ✅ Use transactions for multi-step database operations
- ✅ Validate input before database operations
- ✅ Wrap errors with context: `fmt.Errorf("...: %w", err)`
- ✅ Follow clean architecture layering

## CI/CD

Workflows live in [`.github/workflows/`](.github/workflows/) (monorepo pattern — `paths` filters
select the module):

- `test.yml` — govern library tests (triggers on `config/**`, `http/**`, …)
- `test-sample.yml` — sample app tests (triggers on `examples/golang-sample/**`)
- `push.yml` — Docker image build (context: `examples/golang-sample`)

## File Naming

- **Go files:** `snake_case` (`user_service.go`)
- **Test files:** `source_test.go` (`user_service_test.go`)
- **Packages:** `snake_case`, singular, lowercase (`package user`)
