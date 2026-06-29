# govern/http/echo

Import: `github.com/haipham22/govern/http/echo`

Echo framework integration with govern middleware conventions. Bridges Echo apps with govern logging, request ID, JWT, CORS, Swagger.

## Use When

- App uses Echo framework.
- Need govern-style middleware wiring on Echo.

## Typical Wiring

```go
import (
    "github.com/haipham22/govern/http/echo"
    "github.com/labstack/echo/v5"
)

e := echo.New()
httpEcho.WithEchoSwagger(e,
    httpEcho.WithSwaggerEnabled(true),
)
// attach govern middleware to e
```

## Rules

- ✅ Use govern Echo helpers for middleware (JWT, Swagger, CORS) when available.
- ✅ Map govern/app errors to Echo responses via centralized error handler.
- ✅ Keep Echo handlers thin; delegate to usecase.
- ❌ Do not re-implement middleware that govern/http/echo already wraps.
- ❌ Do not couple handlers to GORM/DB directly.

## Boundary

If app has its own error envelope (e.g. `apperrors`), map govern/Echo errors to app errors at the handler boundary — do not leak framework errors to clients.

## Avoid

- Manual Echo + govern/http bridging.
- Per-handler error formatting instead of centralized handler.

## Reference

Source: [`http/echo/`](../../../../../../../http/echo/). See also [http.md](http.md) for server lifecycle.
