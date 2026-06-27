# golang-sample - Go Development Guide

Production-ready Go API demonstrating clean architecture principles with Echo framework and govern package stack.

## Project Setup

**Prerequisites:**
- mise installed (manages Go version and tools)
- Go 1.26.0 (managed by mise)
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

1. **Update mocks (if added new interfaces):**
   ```bash
   mise exec -- mockery
   ```

2. **Commit with conventional commits:**
   ```bash
   git add .
   git commit -m "feat: add user registration endpoint"
   ```

## Development Rules

**Comprehensive rules are maintained in [`.claude/rules/`](.claude/rules/):**

See [`.claude/rules/README.md`](.claude/rules/README.md) for complete rules overview, including:
- Type system rules (pointers, generics, values)
- Context and concurrency patterns
- Error handling and wrapping
- Input validation with go-playground/validator
- Database operations with GORM
- API documentation with Swagger/OpenAPI
- Testing best practices
- Clean architecture structure
- And more...

**Quick reference - Key rules:**
- ✅ Always use `mise exec --` for Go commands
- ✅ Pass context as first parameter in I/O functions
- ✅ Use transactions for multi-step database operations
- ✅ Validate input before database operations
- ✅ Add Swagger annotations to HTTP handlers
- ✅ Follow clean architecture layering

---

## Architecture

**Clean Architecture based on bxcodec/go-clean-arch pattern.**

See: [`.claude/rules/clean-architecture.md`](.claude/rules/clean-architecture.md) for detailed folder structure rules.

**Quick Reference:**
- **domain/** - Pure entities (flat structure, no interfaces)
- **usecase/** - Business logic + repository interfaces
- **repository/** - Database implementations
- **handler/** - HTTP/gRPC/job/kafka delivery
- **bootstrap/** - Manual dependency injection

**Dependency Rule:** `handler → usecase → domain`, `repository → domain`

---

### File Naming

- **Go files:** `snake_case` (`user_service.go`)
- **Test files:** `source_test.go` (`user_service_test.go`)
- **Packages:** `snake_case` (`package user_service`)

