# Development Guide

Complete guide for developing, testing, and building golang-sample.

## Table of Contents

- [Development Setup](#development-setup)
- [Testing](#testing)
- [Mock Generation](#mock-generation)
- [Pre-commit Hooks](#pre-commit-hooks)
- [Building](#building)
- [Debugging](#debugging)
- [Code Quality](#code-quality)
- [Workflow](#workflow)

## Development Setup

### Install Development Tools

```bash
# Install Wire (dependency injection)
go install github.com/google/wire/cmd/wire@latest

# Install Swag (Swagger documentation)
go install github.com/swaggo/swag/cmd/swag@latest

# Install gomock (mocking framework)
go install github.com/golang/mock/mockgen@latest

# Install pre-commit hooks
pip install pre-commit
```

### Install Pre-commit Hooks

```bash
cd golang-sample
pre-commit install
```

### Configure IDE

**VSCode:**

Install extensions:
- Go (golang.go)
- Swagger Viewer (32bit.shadowbrand.swagger-viewer)

**GoLand:**
- Enable Go Modules integration
- Configure Wire annotation processing

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./pkg/utils/password -v

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage

Current coverage:
- `pkg/utils/password`: **100%** ✅
- `internal/handler/rest/health`: **52.9%** ✅
- `internal/handler/rest/auth`: Mock tests
- Overall: **~60%**

### Run Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./pkg/utils/password

# Run specific benchmark
go test -bench=TestHashPassword -benchmem ./...
```

### Example Test

```go
func TestHashPassword(t *testing.T) {
    pwd := "SecurePassword123!"
    hash, err := password.HashPassword(pwd)

    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.NotEqual(t, pwd, hash)
}
```

## Mock Generation

### Using Gomock

```bash
# Generate mocks for all interfaces
go generate ./...

# Generate mocks for specific package
go generate ./internal/storage

# Regenerate wire dependencies
go generate ./internal
```

### Mock Interface Example

```go
//go:generate mockgen -source=storage.go -destination=mock/mock_storage.go -package=mock

type Storage interface {
    CreateUser(ctx context.Context, user *models.User) (*models.User, error)
    FindUserByUsername(ctx context.Context, username string) (*models.User, error)
}
```

### Using Mocks in Tests

```go
func TestAuthController_Login(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mock.NewMockStorage(ctrl)
    mockStorage.EXPECT().
        FindUserByUsername(gomock.Any(), "testuser").
        Return(expectedUser, nil)

    controller := auth.NewAuthController(mockStorage)
    // Test login...
}
```

## Pre-commit Hooks

### Available Hooks

| Hook | Description |
|------|-------------|
| `go-fmt` | Format Go code |
| `goimports-repo` | Sort imports (repository order) |
| `go-imports-local` | Sort imports (local packages grouped) |
| `gomock` | Generate mocks when go.mod changes |
| `trailing-whitespace` | Remove trailing whitespace |
| `end-of-file-fixer` | Ensure files end with newline |
| `check-yaml` | Validate YAML syntax |
| `check-json` | Validate JSON syntax |
| `check-toml` | Validate TOML syntax |
| `check-merge-conflict` | Detect merge conflicts |
| `check-case-conflict` | Detect case conflicts |
| `detect-private-key` | Detect private keys |
| `mixed-line-ending` | Fix line endings |

### Run Hooks Manually

```bash
# Run all hooks
pre-commit run --all-files

# Run specific hook
pre-commit run go-fmt --all-files

# Run hooks on staged files
pre-commit run
```

### Hook Configuration

Configuration in `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-imports-local
        name: go imports with local sort
        entry: goimports -local golang-sample
        language: system
        types: [ go ]
```

### Skip Hooks (Not Recommended)

```bash
# Skip hooks for specific commit
git commit --no-verify -m "WIP: work in progress"
```

## Building

### Build Binary

```bash
# Build for current platform
go build -o bin/serverd .

# Build with optimizations
go build -ldflags="-s -w" -o bin/serverd .

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/serverd-linux .

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o bin/serverd-mac .

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/serverd.exe .
```

### Build with Version Info

```bash
VERSION=$(git describe --tags --always)
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
go build -ldflags="$LDFLAGS" -o bin/serverd .
```

### Build Docker Image

```bash
# Build image
docker build -t golang-sample:latest .

# Build with tag
docker build -t golang-sample:v1.0.0 .

# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 -t golang-sample:latest .
```

### Run Binary

```bash
# Direct execution
./bin/serverd

# With environment file
export $(cat .env | xargs) && ./bin/serverd

# With flags
./bin/serverd --port 8080 --debug
```

## Debugging

### Enable Debug Mode

```bash
# Set debug environment variable
export APP_DEBUG=true

# Or use flag
./bin/serverd --debug
```

### Using Delve (Go Debugger)

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug application
dlv debug main.go -- serverd

# Debug with specific port
dlv debug --listen=:2345 --headless=true --api-version=2 main.go -- serverd
```

### Remote Debugging (VSCode)

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["serverd"],
      "env": {
        "APP_DEBUG": "true"
      }
    }
  ]
}
```

### Logging

Logs are structured with Zap:

```go
logger.Info("User logged in",
  zap.String("user_id", userID),
  zap.String("ip", c.RealIP()),
)
```

View logs:
```bash
# Follow logs
./bin/serverd 2>&1 | tee server.log

# Filter logs
grep "ERROR" server.log
grep "user_id" server.log
```

## Code Quality

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Run specific linters
golangci-lint run --enable-all

# Run fast linters only
golangci-lint run --fast
```

### Static Analysis

```bash
# Run go vet
go vet ./...

# Run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Run errcheck
go install github.com/kisielk/errcheck@latest
errcheck ./...
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Format with goimports
goimports -w -local golang-sample .

# Fix imports
goimports -local golang-sample -w .
```

### View Coverage in Browser

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Or open automatically
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

## Workflow

### Feature Development

1. **Create branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make changes**
   ```bash
   # Edit code
   vim internal/handler/rest/auth/login.go

   # Run tests
   go test ./internal/handler/rest/auth -v

   # Run pre-commit hooks
   pre-commit run
   ```

3. **Commit changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/amazing-feature
   gh pr create --title "Add amazing feature" --body "Description..."
   ```

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `chore:` Maintenance tasks
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `style:` Code style changes (formatting, etc.)

Examples:
```bash
git commit -m "feat: add user profile endpoint"
git commit -m "fix: correct password verification logic"
git commit -m "docs: update quickstart guide"
git commit -m "test: add integration tests for auth"
git commit -m "refactor: extract validation to separate package"
```

### Pull Request Checklist

- [ ] Tests pass (`go test ./...`)
- [ ] Code formatted (`go fmt ./...`)
- [ ] Pre-commit hooks pass (`pre-commit run --all-files`)
- [ ] Coverage maintained or improved
- [ ] Documentation updated
- [ ] Swagger docs regenerated (if API changed)
- [ ] Wire dependencies regenerated (if DI changed)

### Troubleshooting

#### Wire Generation Fails

```bash
# Clear cache and regenerate
rm internal/wire_gen.go
go generate ./internal
```

#### Tests Fail with "no such file or directory"

```bash
# Sync dependencies
go mod tidy
go mod download
```

#### Import Path Issues

```bash
# Fix import paths
goimports -local golang-sample -w .
```

#### Docker Build Fails

```bash
# Clear Docker cache
docker builder prune -a

# Rebuild without cache
docker build --no-cache -t golang-sample .
```

## Performance Profiling

### CPU Profiling

```bash
# Enable profiling
./bin/serverd --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof

# Visualize
go tool pprof -http=:8081 cpu.prof
```

### Memory Profiling

```bash
# Enable memory profiling
./bin/serverd --memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

## Best Practices

1. **Write tests first** (TDD when practical)
2. **Keep functions small** (<50 lines)
3. **Use interfaces** for dependencies
4. **Handle errors properly** (never ignore errors)
5. **Log with context** (include request IDs, user IDs)
6. **Use structured logging** (Zap, not fmt.Println)
7. **Generate mocks** for all interfaces
8. **Run pre-commit hooks** before committing
9. **Update documentation** when changing APIs
10. **Profile before optimizing** (measure, don't guess)

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Wire User Guide](https://github.com/google/wire/blob/main/docs/guide.md)
- [Echo Guide](https://echo.labstack.com/docs)
- [Govern Package](https://github.com/haipham22/govern)
