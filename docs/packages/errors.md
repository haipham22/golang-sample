# errors

Standardized error handling with error codes and wrapping.

## Overview

The `errors` package provides error handling utilities including error codes, wrapping, and standard error predicates built on `pkg/errors`.

## Key Types

### ErrorCode

```go
type ErrorCode string

const (
    CodeInternal      ErrorCode = "INTERNAL"
    CodeInvalid       ErrorCode = "INVALID"
    CodeNotFound      ErrorCode = "NOT_FOUND"
    CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
    CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
    CodeForbidden     ErrorCode = "FORBIDDEN"
    CodeConflict      ErrorCode = "CONFLICT"
    CodeRateLimit     ErrorCode = "RATE_LIMIT"
)
```

### ErrorWithCode

```go
type ErrorWithCode struct {
    Code ErrorCode
    Err  error
}
```

## Key Functions

### Basic Error Creation

```go
// Create new error
func New(message string) error

// Format error message
func Errorf(format string, args ...interface{}) error

// Join multiple errors
func Join(errs ...error) error
```

### Error Wrapping

```go
// Wrap error with code
func WrapCode(code ErrorCode, err error) error

// Create new error with code
func NewCode(code ErrorCode, message string) error
```

### Error Inspection

```go
// Check if error is target
func Is(err, target error) bool

// Find first error in chain matching target
func As(err error, target interface{}) bool

// Unwrap error
func Unwrap(err error) error

// Get code from error
func GetCode(err error) (ErrorCode, bool)

// Check if error has specific code
func IsCode(err error, code ErrorCode) bool
```

### Predefined Errors

```go
var (
    ErrInternal     = &ErrorWithCode{Code: CodeInternal, Err: New("internal error")}
    ErrInvalid      = &ErrorWithCode{Code: CodeInvalid, Err: New("invalid input")}
    ErrNotFound     = &ErrorWithCode{Code: CodeNotFound, Err: New("resource not found")}
    ErrUnauthorized = &ErrorWithCode{Code: CodeUnauthorized, Err: New("unauthorized")}
)
```

## Usage

### Basic Error Handling

```go
import "github.com/haipham22/govern/errors"

// Create simple error
err := errors.New("something went wrong")

// Format error
err := errors.Errorf("failed to process %s", itemName)

// Join errors
var errs []error
errs = append(errs, err1, err2)
joinedErr := errors.Join(errs...)
```

### Error Codes

```go
// Create error with code
err := errors.NewCode(errors.CodeNotFound, "user not found")

// Wrap existing error with code
dbErr := db.Query(user)
if dbErr != nil {
    return errors.WrapCode(errors.CodeInternal, dbErr)
}

// Check error code
if errors.IsCode(err, errors.CodeNotFound) {
    // Handle not found
}
```

### Error Inspection

```go
// Check error type
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found
}

// Extract error code
if code, ok := errors.GetCode(err); ok {
    fmt.Printf("Error code: %s\n", code)
}

// Unwrap to check underlying error
if errors.Unwrap(err) != nil {
    // Has underlying error
}
```

### Error Chain

```go
// Wrap errors with context
baseErr := errors.New("database connection failed")
wrappedErr := errors.WrapCode(errors.CodeInternal, baseErr)

// Check chain
if errors.Is(wrappedErr, baseErr) {
    // True - baseErr is in chain
}

// Extract from chain
var dbErr *DatabaseError
if errors.As(wrappedErr, &dbErr) {
    // dbErr contains the DatabaseError
}
```

### Common Pattern: Repository Layer

```go
func (r *userRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.WrapCode(errors.CodeNotFound, err)
        }
        return nil, errors.WrapCode(errors.CodeInternal, err)
    }
    return &user, nil
}
```

### Common Pattern: Handler Layer

```go
func (h *handler) GetUser(c echo.Context) error {
    user, err := h.service.FindByID(c.Param("id"))
    if err != nil {
        if errors.IsCode(err, errors.CodeNotFound) {
            return c.JSON(http.StatusNotFound, map[string]string{
                "code": string(errors.CodeNotFound),
                "message": "User not found",
            })
        }
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "code": string(errors.CodeInternal),
            "message": "Internal server error",
        })
    }
    return c.JSON(http.StatusOK, user)
}
```

## Error Messages

The `ErrorWithCode` type formats errors as:

```
[CODE] message
```

Example:
```go
err := errors.NewCode(errors.CodeNotFound, "user not found")
fmt.Println(err.Error())
// Output: [NOT_FOUND] user not found
```

## Best Practices

1. **Use error codes at layer boundaries** - Wrap errors with appropriate codes when moving between layers
2. **Preserve error chains** - Always use `WrapCode` instead of creating new errors
3. **Check error codes, not messages** - Use `IsCode()` for error handling logic
4. **Define domain-specific codes** - Add codes specific to your domain in your application

## References

- [errors/errors.go](../../errors/errors.go) - Error implementation
