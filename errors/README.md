# errors

Structured error handling with error codes.

## Features

- Error codes for categorization (INTERNAL, INVALID, NOT_FOUND, etc.)
- Error wrapping with code preservation
- Helper functions for error inspection
- Predefined common errors
- Wrappers around `github.com/pkg/errors`

## Installation

```bash
go get github.com/haipham22/govern/errors
```

## Usage

### Creating Errors with Codes

```go
import "github.com/haipham22/govern/errors"

// Create a new error with a code
err := errors.NewCode(errors.CodeNotFound, "User not found")

// Wrap an existing error with a code
err = errors.WrapCode(errors.CodeInternal, originalErr)
```

### Checking Error Codes

```go
// Check if error has a specific code
if errors.IsCode(err, errors.CodeNotFound) {
    // Handle not found
}

// Extract the code from an error
if code, ok := errors.GetCode(err); ok {
    fmt.Printf("Error code: %s\n", code)
}
```

### Using Predefined Errors

```go
// Use predefined common errors
return errors.ErrNotFound
return errors.ErrInvalid
return errors.ErrInternal
return errors.ErrUnauthorized
```

## Error Codes

| Code | Description |
|------|-------------|
| `CodeInternal` | Internal server error |
| `CodeInvalid` | Invalid input |
| `CodeNotFound` | Resource not found |
| `CodeAlreadyExists` | Resource already exists |
| `CodeUnauthorized` | Unauthorized access |
| `CodeForbidden` | Forbidden access |
| `CodeConflict` | Conflict error |
| `CodeRateLimit` | Rate limit exceeded |

## API

### Error Creation

| Function | Description |
|----------|-------------|
| `New(message string) error` | Create a new error |
| `Errorf(format string, args ...interface{}) error` | Create a formatted error |
| `NewCode(code ErrorCode, message string) error` | Create error with code |
| `WrapCode(code ErrorCode, err error) error` | Wrap error with code |

### Error Inspection

| Function | Description |
|----------|-------------|
| `Is(err, target error) bool` | Check if errors match |
| `As(err error, target interface{}) bool` | Extract error of type |
| `Unwrap(err error) error` | Unwrap error |
| `Join(errs ...error) error` | Join multiple errors |
| `GetCode(err error) (ErrorCode, bool)` | Extract error code |
| `IsCode(err error, code ErrorCode) bool` | Check error code |

### Predefined Errors

| Variable | Code |
|----------|------|
| `ErrInternal` | CodeInternal |
| `ErrInvalid` | CodeInvalid |
| `ErrNotFound` | CodeNotFound |
| `ErrUnauthorized` | CodeUnauthorized |
