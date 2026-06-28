# Error Handling

The sample app uses a single typed error package, `internal/errors` (package
`apperrors`), replacing the former `github.com/haipham22/govern/errors` dependency.
It is dependency-free and defines the entire error contract for the app.

## Error model

```go
type Code string                 // machine-readable category
type Error struct {              // typed error carrying a Code + cause
    Code Code
    Err  error
    // ...
}
```

### Codes → HTTP status

| Code              | HTTP | When to use                       |
| ----------------- | ---- | --------------------------------- |
| `CodeInvalid`     | 400  | Malformed/invalid request input   |
| `CodeUnauthorized`| 401  | Missing/invalid authentication    |
| `CodeForbidden`   | 403  | Authenticated but not allowed     |
| `CodeNotFound`    | 404  | Resource does not exist           |
| `CodeConflict`    | 409  | Duplicate / state conflict        |
| `CodeAlreadyExists`| 409 | Alias of Conflict                 |
| `CodeRateLimit`   | 429  | Too many requests                 |
| `CodeInternal`    | 500  | Unexpected internal failure       |

`Code.HTTPStatus()` maps a code to its status; unknown codes default to 500.

## Creating errors

```go
// New typed error (no wrapped cause)
err := apperrors.NewCode(apperrors.CodeConflict, "username already exists")

// Wrap an underlying error with a code
err := apperrors.WrapCode(apperrors.CodeInternal, dbErr)   // nil-safe

// Convenience constructors
err := apperrors.NotFound("user")
err := apperrors.InvalidInput("email is required")
err := apperrors.Unauthorized("invalid token")

// Sentinels (work with errors.Is across a wrapped chain)
if errors.Is(err, apperrors.ErrUnauthorized) { ... }
```

## Reading errors

```go
// Extract the code from any error in the chain
code, ok := apperrors.GetCode(err)

// Check a specific code
if apperrors.IsCode(err, apperrors.CodeNotFound) { ... }
```

`GetCode` uses `errors.As` so it traverses `fmt.Errorf("...: %w", ...)` wrappers.

## Layer conventions

- **Repository** (`internal/repository/...`): wrap DB errors with
  `apperrors.WrapCode(apperrors.CodeInternal, err)`; map `gorm.ErrRecordNotFound`
  → `apperrors.CodeNotFound`.
- **Usecase** (`internal/usecase/...`): emit domain errors with `NewCode` /
  convenience constructors (`Conflict`, `Unauthorized`); perform business
  validation before touching the repo.
- **Handler** (`internal/handler/rest/...`): bind errors →
  `WrapCode(CodeInvalid, err)`; validation errors → return `c.Validate(...)`
  unchanged so field details from `apperrors.Validation(...)` are preserved.
  Otherwise return the usecase error unchanged.

## HTTP error response

`internal/handler/rest/handler.go` (`makeHTTPErrorHandler`) is the single error
→ HTTP boundary. It calls `apperrors.Resolve(err, path, requestID)` to get a
`(status, Response)`:

```go
type Response struct {
    Msg       string       `json:"msg"`
    Error     string       `json:"error"`
    Path      string       `json:"path,omitempty"`
    RequestID string       `json:"request_id,omitempty"`
    Errors    []FieldError `json:"errors,omitempty"` // CodeInvalid only
}
```

- `Code.ClientMessage()` returns a sanitized, client-safe message — **5xx never
  leaks internal details**.
- For `CodeInvalid`, `apperrors.Resolve` fills `Errors[]` from field-level
  detail carried by `apperrors.Validation(...)`.
- `apperrors.LogRequestError(log, err, path, status)` logs at the right level
  (conflict → Warn, 5xx → Error+err, 4xx → Warn); conflict raw errors are not
  logged (may leak existence).

### Example responses

```json
// 409 Conflict
{"msg":"Resource already exists","error":"Resource already exists","path":"/api/register"}

// 400 Invalid (validation)
{"msg":"email is required","error":"email is required",
 "errors":[{"property":"email","msg":"email is required"}],"path":"/api/register"}

// 500 Internal (sanitized)
{"msg":"Internal Server Error","error":"Internal Server Error","path":"/api/login"}
```

## Testing

The package is fully unit-tested (`internal/errors/*_test.go`, 93.7% coverage):
formatting, nil safety, unwrap chain, `GetCode`/`IsCode`, sentinel recognition,
HTTP status mapping, `Resolve`, and logging branches.
