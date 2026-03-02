# Code Standards & Development Guidelines

**Last Updated**: 2026-02-26
**Version**: 2.0.0
**Applies To**: All code within golang-sample project

## Table of Contents

- [Core Development Principles](#core-development-principles)
- [File Organization Standards](#file-organization-standards)
- [Naming Conventions](#naming-conventions)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Standards](#testing-standards)
- [Error Handling Patterns](#error-handling-patterns)
- [Security Standards](#security-standards)
- [Documentation Requirements](#documentation-requirements)
- [Git Standards](#git-standards)
- [Performance Guidelines](#performance-guidelines)

---

## Core Development Principles

### Clean Architecture Principles

1. **Dependency Rule**: Dependencies point inward. Inner layers define interfaces, outer layers implement
2. **Interface Segregation**: Small, focused interfaces for specific use cases
3. **Single Responsibility**: Each component has one clear purpose
4. **Dependency Inversion**: Depend on abstractions, not concretions

### Go Best Practices

- **Idiomatic Go**: Follow Go conventions and idioms
- **Simplicity**: Favor simple, straightforward solutions
- **Readability**: Code should be self-documenting
- **Composition**: Prefer composition over inheritance

---

## Architecture Guidelines

### Layer Rules

**HTTP Handlers (`internal/handler/rest/`):**
- MAY use Echo framework
- MAY bind/validate HTTP requests
- MAY return JSON responses
- MUST NOT contain business logic
- MUST call controllers for operations

**Controllers (`internal/handler/rest/controllers/`):**
- MAY accept `echo.Context` (HTTP layer)
- MUST return schema objects (not domain models)
- MAY use Echo framework
- MUST delegate to services
- MUST convert between schemas and service requests
- MUST NOT access config directly (inject via constructor)

**Services (`internal/service/`):**
- MUST contain business logic
- MUST be framework-agnostic (no Echo types)
- MUST accept injected dependencies
- MUST use interfaces for external dependencies (storage)
- MUST accept/return domain models (not schemas, not ORM entities)
- MUST NOT access HTTP layer types

**Domain Models (`internal/model/`):**
- MUST be pure entities with NO external dependencies
- MUST NOT import GORM, Echo, or any framework packages
- MAY contain business logic methods (Validate, CanLogin, etc.)
- MUST be separate from ORM entities
- MUST NOT contain persistence concerns

**Storage (`internal/storage/`):**
- MUST define interfaces for data access
- MUST use GORM for database operations
- MUST return domain models (not ORM entities)
- MUST convert between ORM entities and domain models
- MUST keep password hash separate from domain model

### Dependency Injection

**Rules:**
1. All dependencies injected via constructors
2. No global state in business logic
3. Use Wire for compile-time dependency injection
4. Read config at composition root (wire providers)
5. Pass config values as constructor parameters

**Example:**
```go
// Service layer - config injected
func NewAuthService(
    log *zap.SugaredLogger,
    storage storage.Storage,
    jwtSecret string,  // Injected, not config.ENV.API.Secret
    jwtExpiration time.Duration,
    hasher PasswordHasher,
) AuthService {
    return &serviceImpl{
        log:           log,
        storage:       storage,
        jwtSecret:     jwtSecret,
        jwtExpiration: jwtExpiration,
        hasher:        hasher,
    }
}
```

---

## File Organization Standards

### Directory Structure

```
golang-sample/
├── cmd/                    # Application entry points
│   ├── serverd.go         # API server command
│   └── root.go            # Root command configuration
├── internal/              # Private application code
│   ├── handler/           # HTTP handlers
│   │   └── rest/          # REST API layer
│   │       ├── controllers/  # Request controllers
│   │       ├── middlewares/  # HTTP middlewares
│   │       ├── swagger/      # API documentation
│   │       ├── handler.go    # Server setup
│   │       ├── routes.go     # Route registration
│   │       └── wire.go       # DI setup
│   ├── service/           # Service layer (business logic)
│   │   └── auth/          # Auth service
│   ├── storage/           # Data access layer
│   │   └── user/          # User storage
│   ├── model/             # Domain models (pure)
│   ├── orm/               # ORM models (GORM)
│   ├── schemas/           # Request/response DTOs
│   └── validator/         # Input validation
├── pkg/                   # Public packages
│   ├── config/           # Configuration management
│   └── utils/            # Utility functions
│       └── password/     # Password hashing
└── scripts/              # Build and deployment scripts
```

### File Naming Conventions

**Go Files**:
- Use **lowercase** with **underscores** for multi-word filenames
- Format: `[feature].go` or `[feature]_[component].go`
- Examples:
  - `user.go` (single component)
  - `user_storage.go` (multiple components)
  - `login_handler.go` (specific functionality)

**Directory Names**:
- Use **lowercase** with **underscores** for multi-word directories
- Examples:
  - `middlewares/` (single word)
  - `storage/` (single word)
  - `routes/auth/` (feature-based)

### File Size Management

**Maximum File Size**: 200 lines of code
- Files exceeding 200 lines MUST be refactored
- Exception: Auto-generated files (clearly marked)

**Refactoring Strategies**:

1. **Extract Functions**: Move related functions to separate files
2. **Group by Feature**: Organize by feature rather than layer
3. **Create Utilities**: Extract reusable code to utils package

**Example**:
```
Before:
user_controller.go (350 lines)

After:
user_controller.go (150 lines)    # Core controller logic
user_validation.go (80 lines)     # Validation helpers
user_response.go (70 lines)       # Response formatting
```

---

## Naming Conventions

### Go Naming Standards

**Packages**:
- Use **singular**, **lowercase** names
- Short, concise names (1-2 words)
- Avoid underscore or camelCase
- Examples: `user`, `storage`, `config`, `postgres`

**Constants**:
- Use **camelCase** or **PascalCase** for exported constants
- Examples:
  ```go
  const maxRetries = 3
  const DefaultTimeout = 30 * time.Second
  ```

**Variables**:
- Use **camelCase** for local and package-level variables
- Use **PascalCase** for exported variables
- Examples:
  ```go
  var userCount int
  var DB *gorm.DB
  ```

**Functions**:
- Use **PascalCase** for exported functions
- Use **camelCase** for unexported functions
- Examples:
  ```go
  func CreateUser() {}     // Exported
  func validateEmail() {}  // Unexported
  ```

**Interfaces**:
- Use **PascalCase** ending with `-er` suffix when appropriate
- Examples:
  ```go
  type Storage interface {}
  type Handler interface {}
  type Validator interface {}
  ```

**Structs**:
- Use **PascalCase** for exported structs
- Use **camelCase** for unexported structs
- Examples:
  ```go
  type User struct {}
  type userRepository struct {}  // unexported
  ```

**Struct Fields**:
- Use **PascalCase** for exported fields
- Use **camelCase** for unexported fields
- Examples:
  ```go
  type User struct {
      ID        uint      // Exported
      Username  string    // Exported
      passwordHash string // Unexported
  }
  ```

### Database Naming

**Table Names**:
- Use **plural**, **snake_case** names
- Examples: `users`, `user_sessions`, `blog_posts`

**Column Names**:
- Use **snake_case** names
- Examples: `user_id`, `created_at`, `password_hash`

**Foreign Keys**:
- Format: `[referenced_table]_[referenced_column]`
- Examples: `user_id`, `post_id`, `category_id`

### API Endpoint Naming

**URL Paths**:
- Use **kebab-case** for multi-word paths
- Use **plural nouns** for collections
- Examples:
  ```
  GET    /users
  GET    /users/:id
  POST   /users
  PUT    /users/:id
  DELETE /users/:id
  ```

**Query Parameters**:
- Use **snake_case** for multi-word parameters
- Examples: `?page_size=10&sort_by=created_at`

---

## Code Style Guidelines

### Formatting

**Indentation**:
- Use **tabs** (Go standard)
- Configure editor to use tabs

**Line Length**:
- Preferred: 80-100 characters
- Hard limit: 120 characters
- Break long lines logically

**Whitespace**:
- One blank line between functions
- No trailing whitespace
- Blank line at end of file

**Imports**:
- Group imports into three sections (standard, third-party, local)
- Sort alphabetically within groups
- Examples:
  ```go
  import (
      "context"
      "time"

      "github.com/labstack/echo/v4"
      "go.uber.org/zap"

      "golang-sample/internal/api/storage"
      "golang-sample/pkg/models"
  )
  ```

### Comments and Documentation

**Package Comments**:
- Every package should have a package comment
- Place in a file named `doc.go` or at the top of the main file
- Examples:
  ```go
  // Package storage provides data access layer implementation
  // with repository pattern for clean architecture.
  package storage
  ```

**Exported Functions**:
- All exported functions must have comments
- Include function name, purpose, parameters, and return values
- Examples:
  ```go
  // CreateUser creates a new user in the database.
  // It validates the input, hashes the password, and returns
  // the created user with generated ID.
  //
  // Parameters:
  //   ctx: Context for request cancellation and timeout
  //   user: User object with username, email, and password
  //
  // Returns:
  //   *models.User: Created user with ID populated
  //   error: Error if validation fails or database operation fails
  func (s *storageHandler) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
  ```

**Complex Logic**:
- Comment complex algorithms or non-obvious logic
- Explain WHY, not WHAT
- Examples:
  ```go
  // Check password after hash comparison to avoid timing attacks
  if password.CheckPasswordHash(user.PasswordHash, req.Password) {
      return echo.ErrUnauthorized
  }
  ```

**TODO Comments**:
- Include assignee and date
- Examples:
  ```go
  // TODO(john, 2026-02-09): Implement refresh token flow
  ```

---

## Testing Standards

### Test Organization

**Test File Structure**:
```
internal/api/
├── routes/
│   └── auth/
│       ├── auth.go
│       ├── login.go
│       └── auth_test.go      # Tests for auth package
└── storage/
    ├── storage.go
    └── storage_test.go       # Tests for storage package
```

**Test Naming**:
- Test files: `[source]_test.go`
- Test functions: `Test[FunctionName][Scenario]`
- Examples:
  ```go
  func TestCreateUser_Success(t *testing.T) {}
  func TestCreateUser_DuplicateEmail(t *testing.T) {}
  func TestLogin_InvalidCredentials(t *testing.T) {}
  ```

### Test Coverage

**⚠️ Current State: 0% coverage** (No tests implemented)

**Minimum Coverage Target**: 80%
- Use `go test -cover ./...` to check coverage
- Use `go test -coverprofile=coverage.out` for detailed report

### Test Structure

**Table-Driven Tests**:
```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid format", "invalid", true},
        {"empty string", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Best Practices

1. **Isolation**: Each test should be independent
2. **Clarity**: Test names should describe what they test
3. **Coverage**: Test both success and failure paths
4. **Mocks**: Use interfaces for easy mocking
5. **Cleanup**: Clean up resources in `defer` or `t.Cleanup()`

---

## Error Handling Patterns

### Error Wrapping

**Use fmt.Errorf with %w**:
```go
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

**Custom Error Types**:
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}
```

### Error Handling in Handlers

**Always handle errors**:
```go
func (c *Controller) CreateUser(ctx echo.Context) error {
    var req schemas.UserRegisterRequest
    if err := ctx.Bind(&req); err != nil {
        return errors.NewRequestBindingError(err)
    }

    if err := c.storage.CreateUser(ctx.Request().Context(), &user); err != nil {
        return errors.Wrap(err, errors.ErrInternalServerError, nil)
    }

    return ctx.JSON(http.StatusCreated, schemas.NewResponse(user, http.StatusCreated))
}
```

### Never Ignore Errors

**Bad**:
```go
user, _ := storage.FindUser(id)  // Don't ignore errors
```

**Good**:
```go
user, err := storage.FindUser(id)
if err != nil {
    return err
}
```

---

## Security Standards

### Password Handling

**Never log passwords**:
```go
// BAD
log.Infof("User login: %s with password %s", username, password)

// GOOD
log.Infof("User login: %s", username)
```

**Always hash passwords**:
```go
hashedPassword, err := password.HashPassword(req.Password)
if err != nil {
    return err
}
```

### SQL Injection Prevention

**Use parameterized queries** (GORM handles this):
```go
// GORM automatically parameterizes queries
db.Where("email = ?", email).First(&user)
```

**Never concatenate user input**:
```go
// BAD
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)

// GOOD
db.Where("email = ?", email).First(&user)
```

### Input Validation

**Always validate input**:
```go
if err := c.Validate(req); err != nil {
    return err
}
```

**Sanitize output**:
- Never expose internal errors to clients
- Use generic error messages for security

### Sensitive Data

**Never commit secrets**:
- Add `.env` to `.gitignore`
- Use environment variables for secrets
- Use secret management in production

---

## Documentation Requirements

### Code Documentation

**Self-Documenting Code**:
- Clear variable and function names
- Logical code organization
- Minimal comments needed

**When to Comment**:
- Complex algorithms
- Non-obvious optimizations
- Public API functions
- Configuration options

### API Documentation

**Swagger Annotations**:
```go
// PostLogin godoc
//
//	@Summary	Login user
//	@Description	Authenticate user with username and password
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.LoginRequest	true	"Login request"
//	@Success	200			{object}	schemas.Response
//	@Router		/api/login [post]
func (c *Controller) PostLogin(ctx echo.Context) error {
```

### README Documentation

**Project README must include**:
- Project description
- Quick start guide
- Configuration instructions
- API endpoints
- Links to detailed docs

---

## Git Standards

### Commit Messages

**Format**: Conventional Commits
```
type(scope): description

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks

**Examples**:
```
feat(auth): add JWT authentication

Implements JWT-based authentication with login endpoint.
Includes password hashing and token generation.

Closes #123
```

### Branch Naming

**Format**: `type/description`

**Types**:
- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates

**Examples**:
```
feature/user-authentication
fix/database-connection-timeout
refactor/user-service-cleanup
docs/api-reference-update
```

### Pre-Commit Checklist

- ✅ No secrets or credentials
- ✅ No debug code or console.logs
- ✅ All tests pass
- ✅ Code follows style guidelines
- ✅ No linting errors (`golangci-lint run`)
- ✅ Files under 200 lines
- ✅ Conventional commit message

---

## Performance Guidelines

### Database Operations

**Use context for cancellation**:
```go
user, err := storage.FindUserWithContext(ctx.Request().Context(), username)
```

**Batch operations when possible**:
```go
// BAD: N+1 queries
for _, userID := range userIDs {
    user, _ := storage.FindUser(userID)
}

// GOOD: Single query with IN clause
users, _ := storage.FindUsersByIDs(userIDs)
```

### Memory Management

**Defer cleanup**:
```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

**Reuse buffers**:
```go
var buf bytes.Buffer
for i := 0; i < 1000; i++ {
    buf.Reset()
    // Use buffer
}
```

### Concurrency

**Use goroutines carefully**:
- Always handle goroutine panics
- Use sync.WaitGroup for coordination
- Context for cancellation

**Example**:
```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
wg.Wait()
```

---

## Quality Assurance

### Code Review Checklist

**Functionality**:
- ✅ Implements required features
- ✅ Handles edge cases
- ✅ Error handling complete
- ✅ Input validation present

**Code Quality**:
- ✅ Follows naming conventions
- ✅ Adheres to file size limits
- ✅ DRY principle applied
- ✅ Well-structured and organized

**Security**:
- ✅ No hardcoded secrets
- ✅ Input sanitization
- ✅ Proper error handling
- ✅ Secure dependencies

**Testing**:
- ✅ Unit tests included
- ✅ Edge cases tested
- ✅ Error paths covered

---

## Enforcement

### Automated Checks

**Pre-Commit**:
- golangci-lint for linting
- gofmt for formatting
- go vet for code issues

**Pre-Push**:
- All tests pass
- Coverage threshold met
- Build succeeds

**CI/CD**:
- Full test suite
- Security scanning
- Performance benchmarks

---

## References

### External Standards

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

### Internal Documentation

- [Project Overview PDR](./project-overview-pdr.md)
- [Codebase Summary](./codebase-summary.md)
- [System Architecture](./system-architecture.md)

---

**Document Version**: 2.0.0
**Last Reviewed**: 2026-02-09
**Next Review**: 2026-03-09
**Maintainer**: Development Team
