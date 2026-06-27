# Go Coding Standards

**Essential Go coding standards for consistency and readability.**

---

## File Naming

**Go Files:**
- Use `snake_case` for multi-word filenames
- Format: `[feature].go` or `[feature]_[component].go`
- Examples:
  - `user.go` (single component)
  - `user_storage.go` (multiple components)
  - `login_handler.go` (specific functionality)
  - `auth_service.go` (service layer)

**Test Files:**
- Name: `[source]_test.go`
- Examples: `user_test.go`, `auth_service_test.go`

**Directory Names:**
- Use `lowercase` with optional `underscores` for multi-word
- Examples: `middlewares/`, `storage/`, `routes/auth/`

**Package Names:**
- Use **singular**, **lowercase** names
- Short, concise (1-2 words)
- No underscores or camelCase
- Examples: `user`, `storage`, `config`, `postgres`, `auth`

---

## Naming Conventions

**Packages:**
```go
package user          // ✅ singular, lowercase
package auth          // ✅ singular, lowercase
package user_storage  // ❌ avoid underscores
package User           // ❌ no PascalCase
package users         // ❌ no plural
```

**Constants:**
```go
const maxRetries = 3              // ✅ camelCase (unexported)
const DefaultTimeout = 30 * time.Second  // ✅ PascalCase (exported)
const (
    ErrorCodeInvalid   = "INVALID"      // ✅ PascalCase for exported
    errorNotFound      = "not_found"    // ✅ camelCase for unexported
)
```

**Variables:**
```go
var userCount int               // ✅ camelCase (unexported)
var DB *gorm.DB                 // ✅ PascalCase (exported)
var httpRequestCount int        // ✅ camelCase for local vars
```

**Functions:**
```go
func CreateUser() {}      // ✅ PascalCase (exported)
func validateEmail() {}   // ✅ camelCase (unexported)
func NewAuthService() {} // ✅ Pascal for constructor
```

**Interfaces:**
```go
type Storage interface {}        // ✅ PascalCase
type UserRepository interface {} // ✅ PascalCase, descriptive
type Reader interface {}         // ✅ Ends with -er pattern
```

**Structs:**
```go
type User struct {}                // ✅ PascalCase (exported)
type userRepository struct {}      // ✅ camelCase (unexported)
type HTTPServer struct {}          // ✅ PascalCase (acronym OK)
```

**Struct Fields:**
```go
type User struct {
    ID        uint      // ✅ PascalCase (exported)
    Username  string    // ✅ PascalCase (exported)
    passwordHash string // ✅ camelCase (unexported)
    JWTToken  string    // ✅ Acronyms can be PascalCase
}
```

**Method Receivers:**
```go
// Value receiver - use when method doesn't modify struct
func (u User) Validate() error { ... }

// Pointer receiver - use when method modifies struct
func (u *User) Save() error { ... }

// Receiver name should be:
// - 1-2 characters abbreviating type
// - Consistent across methods
func (s *userService) CreateUser() { ... }  // ✅ "s" for service
func (r *userRepository) Find() { ... }     // ✅ "r" for repository
```

---

## Code Formatting

**Indentation:**
```go
// Use tabs (Go standard)
func example() {
	if true {
		fmt.Println("tabs, not spaces")
	}
}
```

**Line Length:**
- Preferred: 80-100 characters
- Hard limit: 120 characters
- Break long lines logically

**Whitespace:**
```go
// ✅ One blank line between functions
func function1() {}

func function2() {}

// ✅ No trailing whitespace
// ✅ Blank line at end of file
```

**Imports Organization:**
```go
import (
	// Standard library
	"context"
	"time"
	"fmt"

	// Third-party packages
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	// Local packages
	"github.com/haipham22/govern/examples/golang-sample/internal/handler"
	"github.com/haipham22/golang-sample/internal/usecase"
)
```

---

## File Size Management

**Maximum File Size:** 200 lines
- Files exceeding 200 lines MUST be refactored
- Exception: Auto-generated files (mark clearly)

**Refactoring Strategies:**
```go
// Before: user_controller.go (350 lines)

// After:
user_controller.go       (150 lines)  // Core controller logic
user_validation.go        (80 lines)  // Validation helpers
user_response.go          (70 lines)  // Response formatting
```

**File Naming After Refactoring:**
- Use descriptive, long names if needed
- Self-documenting is better than short obscure names
- Examples: `auth_controller_validation.go` is OK

---

## Struct Organization

**Field Ordering:**
```go
type User struct {
	// 1. Exported fields first
	ID       uint
	Username string
	Email    string

	// 2. Unexported fields last
	passwordHash string
	createdAt    time.Time
}
```

**Tag Ordering:**
```go
type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Username  string    `json:"username" gorm:"unique"`
	Password  string    `json:"-" gorm:"column:password_hash"` // json:"-" to hide
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
```

---

## Function Organization

**Exported Functions First:**
```go
package auth

// Exported functions first
func NewService() *Service { ... }
func (s *Service) Login() error { ... }

// Unexported functions last
func validateEmail() error { ... }
func hashPassword() string { ... }
```

**Function Length:**
- Preferred: ≤50 lines
- Hard limit: 100 lines
- Break long functions into smaller helpers

---

## Error Handling

**Never Ignore Errors:**
```go
// ❌ BAD
user, _ := storage.FindUser(id)

// ✅ GOOD
user, err := storage.FindUser(id)
if err != nil {
	return err
}
```

**Error Wrapping:**
```go
// ✅ Wrap with context
if err != nil {
	return fmt.Errorf("failed to create user: %w", err)
}

// ✅ Custom error types
if err != nil {
	return &ValidationError{
		Field:   "email",
		Message: "invalid email format",
	}
}
```

---

## Comments

**Package Comments:**
```go
// Package auth provides authentication and authorization services
// including JWT token generation and password hashing.
package auth
```

**Exported Functions:**
```go
// CreateUser creates a new user in the database.
// It validates the input, hashes the password, and returns
// the created user with generated ID.
//
// Parameters:
//   ctx: Context for request cancellation and timeout
//   req: CreateUserRequest with username, email, and password
//
// Returns:
//   *User: Created user with ID populated
//   error: Error if validation fails or database operation fails
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
```

**Complex Logic:**
```go
// Check password after hash comparison to avoid timing attacks
if subtle.ConstantTimeCompare([]byte(hashed), []byte(req.Password)) != 1 {
	return ErrInvalidCredentials
}
```

---

## Interface Rules

**Define by Consumer:**
```go
// ✅ GOOD: Interface defined by usecase (consumer)
package auth

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, user *User) error
}

// ✅ Repository implements interface
package postgres

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	// implementation
}
```

---

## Constants and Enums

**Use iota for Enums:**
```go
type UserRole int

const (
	RoleGuest UserRole = iota
	RoleUser
	RoleAdmin
)

func (r UserRole) String() string {
	switch r {
	case RoleGuest:
		return "guest"
	case RoleUser:
		return "user"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}
```

---

## Export Rules

**Exported vs Unexported:**
```go
// Exported (PascalCase) - visible outside package
type User struct {}
func NewUser() *User {}
const MaxRetries = 3

// Unexported (camelCase) - package-private
type userConfig struct {}
func validateUser() error {}
const maxRetries = 3
```

**When to Export:**
- ✅ Export: Types, functions used by other packages
- ✅ Export: Constants used as configuration
- ❌ Unexport: Internal helpers, implementation details
- ❌ Unexport: Types only used within package

---

## Acronyms in Names

**Common Acronyms:**
- HTTP → `httpServer` (not `HTTPServer` for unexported)
- URL → `urlPath` (not `URLPath`)
- ID → `userID` (not `UserID` for field)
- JSON → `jsonTags` (not `JSONTags` for unexported)
- JWT → `jwtToken` (not `JWTToken` for unexported)
- DB → `dbConnection` (not `DBConnection` for unexported)

**Exported Acronyms:**
```go
type HTTPServer struct {}  // ✅ OK for exported type
func ParseURL() {}         // ✅ OK for exported function
```

**Unexported Acronyms:**
```go
type httpServer struct {}  // ✅ Use camelCase for unexported
func parseURL() {}         // ✅ Use camelCase for unexported
```

---

## Testing Conventions

**Test Function Names:**
```go
func TestCreateUser_Success(t *testing.T) {}
func TestCreateUser_DuplicateEmail(t *testing.T) {}
func TestCreateUser_InvalidInput(t *testing.T) {}
```

**Table-Driven Tests:**
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

---

## Godoc Conventions

**Package Documentation:**
```go
// Package auth provides authentication and authorization services.
//
// Features:
//   - JWT token generation and validation
//   - Password hashing using bcrypt
//   - User session management
//
// Usage:
//   service := auth.NewService(repo, logger)
//   token, err := service.Login(ctx, req)
package auth
```

**Function Documentation:**
```go
// Login authenticates a user and returns a JWT token.
//
// The token is valid for 24 hours and includes user claims
// for authorization checks.
//
// Example:
//   token, err := service.Login(ctx, LoginRequest{
//       Email: "user@example.com",
//       Password: "password",
//   })
func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error)
```

---

## Quality Checks

**Pre-Commit:**
- ✅ File names use snake_case
- ✅ Package names are lowercase, singular
- ✅ Functions follow PascalCase/camelCase
- ✅ Struct fields use PascalCase/camelCase
- ✅ Files ≤200 lines
- ✅ No unused imports
- ✅ No unused variables

**Pre-Push:**
- ✅ All tests pass
- ✅ Code follows naming conventions
- ✅ Comments present for exported functions
- ✅ Error handling complete

---

## Quick Reference

| Category | Convention | Example |
|----------|-----------|---------|
| Files | snake_case | `user_service.go` |
| Packages | lowercase, singular | `package auth` |
| Constants (exported) | PascalCase | `const MaxRetries = 3` |
| Constants (unexported) | camelCase | `const maxRetries = 3` |
| Variables (exported) | PascalCase | `var DB *gorm.DB` |
| Variables (unexported) | camelCase | `var userCount int` |
| Functions (exported) | PascalCase | `func CreateUser()` |
| Functions (unexported) | camelCase | `func validateEmail()` |
| Interfaces | PascalCase | `type UserRepository` |
| Structs (exported) | PascalCase | `type User struct` |
| Structs (unexported) | camelCase | `type userRepository` |
| Fields (exported) | PascalCase | `ID string` |
| Fields (unexported) | camelCase | `passwordHash string` |
| Method receivers | 1-2 chars | `(s *Service)` |

---

## References

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
