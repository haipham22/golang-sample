---
title: "Phase 09: Custom Error Types Implementation"
description: "Create custom error types package to replace govern/errors"
status: pending
priority: P1
effort: 3h
dependsOn: [phase-08-setup-wire-removal.md]
---

## Overview

**Priority**: P1 | **Status**: pending | **Effort**: 3h

Create `internal/errors` package with custom error types that match govern/errors functionality while following Go 1.25+ best practices.

**Working Directory**: All operations in this phase are performed in the `examples/golang-sample/` directory. All file paths are relative to `examples/golang-sample/`.

## Context Links

- **Parent Plan**: [plan.md](./plan.md)
- **Previous Phase**: [phase-08-setup-wire-removal.md](./phase-08-setup-wire-removal.md)
- **Next Phase**: [phase-10-centralized-error-management.md](./phase-10-centralized-error-management.md)
- **Related Files**: `examples/golang-sample/internal/model/errors.go`, `examples/golang-sample/internal/handler/rest/handler.go`

## Key Insights

**govern/errors Analysis**:
- 6 error codes: CodeInvalid, CodeNotFound, CodeUnauthorized, CodeForbidden, CodeConflict, CodeInternal
- Error wrapping with `WrapCode()`
- Code extraction with `GetCode()`
- HTTP status mapping already in handler.go

**Go Best Practices**:
- Use error wrapping with `%w` for error chain preservation
- Implement `errors.Is()` and `errors.As()` support
- Custom error types should implement `Error() string`
- Separate error types for different error categories

## Requirements

### Functional Requirements
1. Create custom error types matching govern/errors functionality
2. Support error wrapping with context preservation
3. Implement error code system (replace govern/errors codes)
4. Add error creation helper functions
5. Support errors.Is/As for error checking

### Non-Functional Requirements
- Zero breaking changes to error handling behavior
- Maintain same HTTP status code mappings
- Preserve error logging behavior
- Add comprehensive unit tests (≥80% coverage)

## Architecture

**Error Type Hierarchy with Envelope Pattern:**
```
internal/errors/
├── errors.go              # Core error types (AppError) + envelope types
├── codes.go               # Error code definitions
├── wrap.go                # Error wrapping functions with envelope
├── envelope/             # Error envelope types (NEW)
│   ├── db_error.go         # Database error envelope
│   ├── config_error.go     # Configuration error envelope
│   ├── logger_error.go     # Logger error envelope
│   └── http_error.go      # HTTP error envelope
├── errors_test.go         # Unit tests
└── doc.go                 # Package documentation
```

**Error Code Mapping**:
```
govern/errors → Custom
CodeInvalid    → ErrCodeInvalid (400)
CodeNotFound   → ErrCodeNotFound (404)
CodeUnauthorized → ErrCodeUnauthorized (401)
CodeForbidden  → ErrCodeForbidden (403)
CodeConflict   → ErrCodeConflict (409)
CodeInternal   → ErrCodeInternal (500)
```

**Error Wrapping Flow with Envelope**:
```
Original error → WrapCode() → AppError + Envelope → errors.Is/As support
```

### Error Envelope Patterns

**Database Error Envelope:**
```go
// internal/errors/envelope/db_error.go
package envelope

import "path/to/internal/errors"

type DBError struct {
    Op       string // Operation: "find_by_id", "create", "update"
    Table    string // Table: "users", "products", "orders"
    Err      error  // Underlying error
    Severity string // "transient" or "permanent"
}

func (e *DBError) Error() string {
    return fmt.Sprintf("%s on %s: %v", e.Op, e.Table, e.Err)
}

func (e *DBError) Unwrap() error {
    return e.Err
}

// Usage in repository
func (r *postgresUserRepository) FindByID(ctx context.Context, id int64) (User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return User{}, &errors.AppError{
                Code: errors.ErrCodeNotFound,
                Err: &envelope.DBError{
                    Op:    "find_by_id",
                    Table: "users",
                    Err:   err,
                },
                Message: fmt.Sprintf("user with ID %d not found", id),
            }
        }
        return User{}, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &envelope.DBError{
                Op:       "find_by_id",
                Table:    "users",
                Err:      err,
                Severity: "transient",
            },
            Message: "database error",
        }
    }
    return user, nil
}
```

**Config Error Envelope:**
```go
// internal/errors/envelope/config_error.go
package envelope

type ConfigError struct {
    Key   string // Config key that failed
    Value string // Invalid value
    Err   error  // Underlying error
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config key %s=%s: %v", e.Key, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
    return e.Err
}

// Usage in config loading
func LoadConfig(path string) (*Config, error) {
    if err := viper.ReadInConfig(); err != nil {
        return nil, &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &envelope.ConfigError{
                Key: "config_file",
                Err: err,
            },
            Message: "failed to read config",
        }
    }
    // ...
}
```

**Logger Error Envelope:**
```go
// internal/errors/envelope/logger_error.go
package envelope

type LoggerError struct {
    Op  string // Operation: "create_logger", "write_log"
    Err error  // Underlying error
}

func (e *LoggerError) Error() string {
    return fmt.Sprintf("logger %s: %v", e.Op, e.Err)
}

func (e *LoggerError) Unwrap() error {
    return e.Err
}

// Usage in logger setup
func NewLogger(cfg LogConfig) (*zap.Logger, error) {
    logger, err := config.Build()
    if err != nil {
        return nil, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &envelope.LoggerError{
                Op:  "create_logger",
                Err: err,
            },
            Message: "failed to create logger",
        }
    }
    return logger, nil
}
```

**HTTP Error Envelope:**
```go
// internal/errors/envelope/http_error.go
package envelope

import (
    "net/http"
    "path/to/internal/errors"
)

type HTTPError struct {
    Method     string // HTTP method
    Path       string // Request path
    StatusCode int    // HTTP status code
    Err        error  // Underlying error
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("%s %s: %d - %v", e.Method, e.Path, e.StatusCode, e.Err)
}

func (e *HTTPError) Unwrap() error {
    return e.Err
}

// Usage in HTTP handler
func handleError(c echo.Context, err error) error {
    var httpErr *envelope.HTTPError
    if errors.As(err, &httpErr) {
        // Already wrapped with HTTP context
        return err
    }
    
    // Wrap with HTTP context
    return &envelope.HTTPError{
        Method:     c.Request().Method,
        Path:       c.Request().URL.Path,
        StatusCode: getStatusCode(err),
        Err:        err,
    }
}

func getStatusCode(err error) int {
    if appErr, ok := err.(*errors.AppError); ok {
        return appErr.Code.HTTPStatus()
    }
    return http.StatusInternalServerError
}
```

**Error Envelope Benefits:**
- ✅ Structured error information (operation, table, severity)
- ✅ Consistent error format across layers
- ✅ Easy to classify errors (transient vs permanent)
- ✅ Better debugging with operation context
- ✅ Proper error wrapping chain
- ✅ Supports retry logic based on severity

## Related Code Files

### Files to Create
- `internal/errors/errors.go` - Core error types
- `internal/errors/codes.go` - Error code definitions
- `internal/errors/wrap.go` - Error wrapping functions
- `internal/errors/errors_test.go` - Unit tests
- `internal/errors/doc.go` - Package documentation

### Files to Modify
- `internal/model/errors.go` - Replace with custom types

### Files to Delete
- None (phase 03 will replace govern/errors usage)

## Implementation Steps

1. **Create Error Code System** (45m)
   ```go
   // internal/errors/codes.go
   type ErrorCode string

   const (
       ErrCodeInvalid     ErrorCode = "INVALID"
       ErrCodeNotFound    ErrorCode = "NOT_FOUND"
       ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
       ErrCodeForbidden   ErrorCode = "FORBIDDEN"
       ErrCodeConflict    ErrorCode = "CONFLICT"
       ErrCodeInternal    ErrorCode = "INTERNAL"
   )

   func (c ErrorCode) HTTPStatus() int {
       switch c {
       case ErrCodeInvalid: return 400
       case ErrCodeNotFound: return 404
       case ErrCodeUnauthorized: return 401
       case ErrCodeForbidden: return 403
       case ErrCodeConflict: return 409
       default: return 500
       }
   }
   ```

2. **Create Custom Error Type** (60m)
   ```go
   // internal/errors/errors.go
   type AppError struct {
       Code    ErrorCode
       Err     error
       Message string
       Context map[string]interface{}
   }

   func (e *AppError) Error() string
   func (e *AppError) Unwrap() error
   func (e *AppError) Is(target error) bool
   ```

3. **Create Error Wrapping Functions** (45m)
   ```go
   // internal/errors/wrap.go
   func WrapCode(code ErrorCode, err error) *AppError
   func NewCode(code ErrorCode, msg string) *AppError
   func GetCode(err error) (ErrorCode, bool)
   ```

4. **Write Unit Tests** (30m)
   ```go
   // internal/errors/errors_test.go
   - Test error wrapping preserves original error
   - Test errors.Is() works correctly
   - Test errors.As() works correctly
   - Test HTTP status mapping
   - Test error code extraction
   ```

## Todo List

- [x] Create internal/errors package directory
- [x] Implement error code system (codes.go)
- [x] Implement AppError type (errors.go)
- [x] Implement error wrapping functions (wrap.go)
- [x] Write comprehensive unit tests
- [x] Add package documentation (doc.go)
- [x] Verify tests pass (go test ./internal/errors/...)
- [x] Validate error wrapping preserves context
- [x] Test errors.Is/As functionality
- [x] Document error type usage examples

## Success Criteria

**Definition of Done**:
- Custom error types implemented with full functionality
- Error wrapping preserves original error (errors.Unwrap)
- errors.Is/As support working correctly
- HTTP status code mapping matches govern/errors
- Unit tests passing with ≥80% coverage
- Package documentation complete

**Validation Methods**:
```bash
# Run tests
go test ./internal/errors/... -v -cover

# Test error wrapping
go run -exec main.go # Test with sample error

# Validate errors.Is/As
# Should work like govern/errors
```

**Compatibility Check**:
```go
// Old (govern/errors)
err := governerrors.WrapCode(governerrors.CodeInvalid, err)

// New (custom)
err := errors.WrapCode(errors.ErrCodeInvalid, err)

// Should produce same HTTP response
```

## Risk Assessment

**Potential Issues**:
1. **Error Context Loss**: Wrapping may lose error context
   - Mitigation: Preserve original error with Unwrap(), test thoroughly
2. **HTTP Status Mismatch**: Status codes may differ from govern/errors
   - Mitigation: Unit tests verify exact status mapping
3. **errors.Is/As Issues**: Custom type may not work correctly
   - Mitigation: Follow Go stdlib patterns, test with stdlib errors

**Medium Risk**: This is core infrastructure - errors must work correctly

**Rollback**: If issues found, can delay Phase 03 until fixed

## Security Considerations

- Error messages must not leak sensitive information
- Maintain current security level (no internal details in 5xx errors)
- Validate error context doesn't contain credentials

## Next Steps

**Dependencies**: Phase 01 must be complete

**Follow-up Tasks**:
- Phase 03: Replace govern/errors usage with custom errors
- Phase 04: Manual DI implementation (requires custom errors)

**Transition Criteria**:
- All tests passing → Start Phase 03
- Error wrapping validated → Safe to replace govern/errors
