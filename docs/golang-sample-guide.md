# golang-sample Guide

Complete guide to the golang-sample application demonstrating govern library usage with clean architecture principles.

## Overview

**Module**: `github.com/haipham22/golang-sample`
**Purpose**: Demonstrate govern library integration in a production-ready application
**Architecture**: Clean architecture (Handler → Controller → Service → Storage)

## Quick Start

```bash
# Navigate to sample app
cd examples/golang-sample

# Install dependencies & tools
mise install

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Run the server
make run

# Run tests
make test

# Build binary
make build
```

The API will be available at `http://localhost:8080`

## External Import Configuration

This sample imports govern as an **external dependency** using a `replace` directive in `go.mod`:

```go
module github.com/haipham22/golang-sample

go 1.26.0

require github.com/haipham22/govern v0.0.0

replace github.com/haipham22/govern => ../../
```

**How it works:**
- The `require` directive declares the govern dependency
- The `replace` directive points to the local govern module for development
- In production, remove the `replace` directive and use the published govern version

**Benefits:**
- Tests the govern library as an external consumer would use it
- No `go.work` file needed - uses standard Go module mechanics
- Easy to switch between local and published versions

## Architecture

### Clean Architecture Layers

```
HTTP Request → Handler → Controller → Service → Domain Model ← Storage
```

### Layer Responsibilities

| Layer        | Location                           | Responsibility                          | Govern Usage           |
| ------------ | ---------------------------------- | --------------------------------------- | ---------------------- |
| Entry Point  | `main.go` + `cmd/`               | Bootstrap, signal handling              | `graceful`, `log`      |
| HTTP Handler | `internal/handler/rest/`         | Echo binding, routing, response mapping | `http/echo`, `http/jwt` |
| Controller   | `internal/handler/rest/controllers/` | Request orchestration, error mapping  | `errors`               |
| Service      | `internal/service/auth/`         | Business logic, JWT, password hashing   | N/A (pure business)    |
| Storage      | `internal/storage/user/`         | Data access interface + GORM conversion | `database/postgres`     |
| Model        | `internal/model/`                | Pure domain entities                    | N/A                    |
| ORM          | `internal/orm/`                  | GORM entities                           | `database/postgres`     |
| Schemas      | `internal/schemas/`               | Request/response DTOs                   | N/A                    |
| Validator    | `internal/validator/`            | go-playground/validator wrapper         | N/A                    |
| Config       | `pkg/config/`, `pkg/postgres/`    | Environment config, DB connection       | `config`, `log`         |

### Dependency Rule

**Dependency flow:** `handler → service → model ← storage`

**Key principles:**
- No HTTP→ORM or HTTP→Model direct dependencies
- Always convert through schemas at the HTTP boundary
- Service layer depends on storage interfaces, not implementations
- Model layer has no external dependencies

## Govern Integration

### Configuration Loading

```go
import "github.com/haipham22/govern/config"

// Load configuration from YAML with ENV overrides
cfg, err := config.Load[Config]("config.yaml",
    config.WithEnvFile(".env"),
    config.WithENVPrefix("APP"),
)
```

### Logging

```go
import "github.com/haipham22/govern/log"

// Create logger
logger := log.New(
    log.WithLevel(zapcore.InfoLevel),
    log.WithEncoding("json"),
)
```

### Database Connection

```go
import "github.com/haipham22/govern/database/postgres"

// Connect to PostgreSQL
db, cleanup, err := postgres.New(cfg.PostgresDSN,
    postgres.WithDebug(cfg.Debug),
)
defer cleanup()
```

### HTTP Server

```go
import "github.com/haipham22/govern/http"
import "github.com/haipham22/govern/http/echo"
import "github.com/haipham22/govern/http/middleware"

// Create Echo instance
e := echo.New()

// Add JWT middleware
jwtConfig := &echo.JWTMiddlewareConfig{
    Config:         echo.DefaultConfig(),
    TokenExtractor: echo.DefaultTokenExtractor,
    SkipPaths:      []string{"/health", "/login"},
}
jwtConfig.Config.Secret = cfg.APISecret
e.Use(echo.JWTMiddleware(jwtConfig))

// Add middleware
e.Use(middleware.RequestLog(logger))
e.Use(middleware.Recovery(logger))

// Create server with graceful shutdown
server := http.NewServer(":8080", e, http.WithLogger(logger))
```

### Graceful Shutdown

```go
import "github.com/haipham22/govern/graceful"

// Run with graceful shutdown
graceful.Run(ctx, logger, 30*time.Second, server)
```

## Project Structure

```
examples/golang-sample/
├── cmd/                        # Cobra CLI entry points
│   └── root.go                 # Root command
├── internal/
│   ├── handler/rest/           # HTTP handlers
│   │   ├── controllers/        # Request controllers
│   │   ├── wire.go            # Wire DI setup
│   │   └── routes.go          # Route definitions
│   ├── service/               # Business logic
│   │   └── auth/              # Authentication service
│   ├── storage/               # Data access interfaces
│   │   └── user/              # User storage
│   ├── model/                 # Domain entities
│   │   └── user.go            # User model
│   ├── orm/                   # GORM entities
│   │   └── user.go            # User ORM
│   ├── schemas/               # Request/response DTOs
│   │   └── auth.go            # Auth schemas
│   ├── validator/             # Input validation
│   │   └── validator.go       # Validator wrapper
│   └── mocks/                 # Generated mocks
│       ├── service/           # Service mocks
│       └── storage/           # Storage mocks
├── pkg/                       # Public packages
│   ├── config/                # Configuration
│   └── postgres/              # Database setup
├── go.mod                     # Module definition
├── go.sum                     # Dependencies
├── main.go                    # Entry point
├── Makefile                   # Build commands
├── .env.example               # Environment template
└── config.yaml                # Configuration file
```

## Build, Run, Test Commands

### Build

```bash
# Build binary to bin/serverd
make build

# Or directly
mise exec -- go build -o bin/serverd .
```

### Run

```bash
# Run server
make run

# Or directly
./bin/serverd

# Or with go run
mise exec -- go run . serve
```

### Test

```bash
# Run all tests
make test

# Run with race detector
mise exec -- go test -race ./...

# Run with coverage
mise exec -- go test -cover ./...

# Run specific test
mise exec -- go test ./internal/service/auth/
```

### Lint

```bash
# Run all linters
make lint

# Or individually
mise exec -- goimports -w .
mise exec -- golangci-lint run
mise exec -- staticcheck ./...
mise exec -- errcheck -blank ./...
```

### Generate Mocks

```bash
# Generate mocks with mockery
make generate-mocks

# Or directly
mise exec -- mockery
```

## Testing Approach

### Test Database

The sample app uses **SQLite in-memory** for tests:

```go
// In test setup
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
db.AutoMigrate(&orm.User{})
```

**Benefits:**
- No external database required for tests
- Fast test execution
- Isolated test runs
- Easy CI/CD integration

### Mock Generation

Uses **Mockery** for generating mock implementations:

```bash
# Generate all mocks
mise exec -- mockery

# Generate specific mock
mise exec -- mockery --name=UserStorage
```

**Mock files:** `internal/mocks/service/`, `internal/mocks/storage/`

### Test Structure

```go
func TestAuthService_Register(t *testing.T) {
    // Setup
    mockStorage := new(mocks.UserStorage)
    service := auth.NewService(mockStorage, logger)

    // Test
    mockStorage.On("Create", mock.Anything, mock.AnythingOfType("*orm.User")).
        Return(nil)

    err := service.Register(context.Background(), validRequest)

    // Assert
    assert.NoError(t, err)
    mockStorage.AssertExpectations(t)
}
```

### Test Environment

Uses test-specific configuration:
- `.test-env` - Test environment variables
- `config.test.yaml` - Test configuration

**Helper function:**
```go
func getProjectRoot() string {
    // Returns project root for test config loading
}
```

## Configuration

### Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

**Key variables:**

| Variable            | Description                          | Required |
| ------------------- | ------------------------------------ | -------- |
| `APP_ENV`           | `development` / `production`        | Yes      |
| `APP_DEBUG`         | Enable debug logging                 | No       |
| `APP_POSTGRES_DSN`  | PostgreSQL connection string          | Yes      |
| `APP_API_SECRET`    | JWT signing secret (32+ characters)  | Yes      |

### Example .env

```bash
# Application
APP_ENV=development
APP_DEBUG=true

# Database
APP_POSTGRES_DSN=host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable

# JWT
APP_API_SECRET=your-32-character-secret-key-here
```

## JWT Authentication

### Middleware Setup

```go
import "github.com/haipham22/govern/http/echo"

jwtConfig := &echo.JWTMiddlewareConfig{
    Config:         echo.DefaultConfig(),
    TokenExtractor: echo.DefaultTokenExtractor,
    SkipPaths:      []string{"/health", "/login"},
}
jwtConfig.Config.Secret = cfg.APISecret
e.Use(echo.JWTMiddleware(jwtConfig))
```

### Getting Current User

```go
func (c *AuthController) GetProfile(echoCtx echo.Context) error {
    // Get claims from context
    claims, ok := echo.GetCurrentUser(echoCtx)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
    }

    // Use claims
    return echoCtx.JSON(http.StatusOK, claims)
}
```

### Password Hashing

Uses **bcrypt** for password hashing:

```go
// Hash password
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Compare password
err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
```

## Common Patterns

### Request Flow

```
1. HTTP Request → Handler
2. Handler binds request to schema
3. Handler calls controller
4. Controller validates input
5. Controller calls service
6. Service uses storage interface
7. Storage implementation uses GORM
8. Response flows back through layers
```

### Error Handling

```go
// Storage layer
return errors.New("user not found")

// Service layer wraps with context
return fmt.Errorf("failed to find user: %w", err)

// Controller maps to HTTP status
if errors.Is(err, ErrUserNotFound) {
    return echo.NewHTTPError(http.StatusNotFound, "User not found")
}
```

### Validation

```go
// In controller
if err := v.validator.Struct(req); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
}

// In service
if req.Password != req.ConfirmPassword {
    return ErrPasswordMismatch
}
```

## Best Practices

1. **Always use context** - Pass `context.Context` as first parameter in I/O functions
2. **Use transactions** - For multi-step database operations
3. **Validate early** - Validate at handler/controller layer before business logic
4. **Wrap errors** - Add context to errors with `fmt.Errorf("operation: %w", err)`
5. **Never expose passwords** - Use `json:"-"` tag on password fields
6. **Use interfaces** - Service depends on storage interfaces, not implementations
7. **Mock dependencies** - Use mockery-generated mocks for testing
8. **Keep layers pure** - HTTP logic in handlers, business logic in services

## References

- [Sample App README](../../examples/golang-sample/README.md) - Sample app documentation
- [Sample App CLAUDE.md](../../examples/golang-sample/CLAUDE.md) - Development rules
- [Govern Packages](./packages/) - Package documentation
- [Development Guide](./development.md) - Testing, building, workflow
- [Code Standards](./code-standards.md) - Naming, style, best practices
