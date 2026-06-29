# govern/errors

Import: `github.com/haipham22/govern/errors`

Structured errors with codes for categorization. Wraps `github.com/pkg/errors`.

## Use When

- Working inside root govern library.
- App explicitly chose govern errors as its envelope.

## App Exception

If app has its own errors package (e.g. `internal/errors`, package `apperrors`), use that for app request flow. Do not force `govern/errors` into app. The sample app migration deliberately moved off `govern/errors`.

## Creating Errors

```go
import "github.com/haipham22/govern/errors"

err := errors.NewCode(errors.CodeNotFound, "user not found")
err = errors.WrapCode(errors.CodeInternal, originalErr)
```

## Checking

```go
if errors.IsCode(err, errors.CodeNotFound) { ... }
if code, ok := errors.GetCode(err); ok { fmt.Println(code) }
```

## Error Codes

| Code | Meaning |
|---|---|
| `CodeInternal` | Internal server error |
| `CodeInvalid` | Invalid input |
| `CodeNotFound` | Resource not found |
| `CodeAlreadyExists` | Resource exists |
| `CodeUnauthorized` | Unauthorized |
| `CodeForbidden` | Forbidden |
| `CodeConflict` | Conflict |
| `CodeRateLimit` | Rate limit exceeded |

## API

| Function | Description |
|---|---|
| `New(msg)` | New error |
| `Errorf(fmt, args...)` | Formatted error |
| `NewCode(code, msg)` | Error with code |
| `WrapCode(code, err)` | Wrap with code |
| `Is(err, target)` / `As(err, target)` | Inspection |
| `Unwrap(err)` / `Join(errs...)` | Unwrap/join |
| `GetCode(err)` | Extract code |
| `IsCode(err, code)` | Check code |

Predefined: `ErrInternal`, `ErrInvalid`, `ErrNotFound`, `ErrUnauthorized`.

## Rules

- ✅ Attach codes to client-facing errors so HTTP handler can map status.
- ✅ Wrap with context: `errors.WrapCode(code, err)`.
- ✅ Map raw DB/infra errors to coded errors before crossing layer boundary.
- ❌ Do not use bare `errors.New` for errors needing HTTP status mapping.
- ❌ Do not pull `github.com/pkg/errors` separately — govern already wraps it.

## Avoid

- Losing categorization (bare errors → handler returns 500).
- Mixing govern errors into app that has its own envelope.

## Reference

Source: [`errors/`](../../../../../../../errors/).
