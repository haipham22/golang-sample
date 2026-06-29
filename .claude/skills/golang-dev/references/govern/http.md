# govern/http

Import: `github.com/haipham22/govern/http`

Production HTTP server: graceful shutdown, middleware chain, sensible timeout defaults, logging integration. `net/http`-compatible.

## Defaults

- Read timeout: 10s
- Write timeout: 10s
- Idle timeout: 60s
- Graceful shutdown configurable

## Use When

- App needs HTTP server with middleware + graceful shutdown.
- Want lifecycle integrated with `govern/graceful`.

## Basic Server

```go
import (
    "context"
    "github.com/haipham22/govern/http"
)

handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
})

server := http.NewServer(":8080", handler)
server.Start(context.Background())
```

## With Middleware

```go
import (
    "github.com/haipham22/govern/http"
    "github.com/haipham22/govern/http/middleware"
)

server := http.NewServer(":8080", handler,
    http.WithRequestID(),
    http.WithLogging(logger),
    http.WithRecovery(logger),
    http.WithCORS(&middleware.CORSConfig{
        AllowedOrigins: []string{"*"},
    }),
)
```

## Custom Middleware

```go
server := http.NewServer(":8080", handler)
server.Use(authMiddleware)
server.Start(context.Background())
```

## Middleware Chain (Reusable)

```go
chain := http.NewChain(loggingMiddleware, authMiddleware, recoveryMiddleware)
handler := chain.Then(finalHandler)
```

## Common Options

| Option | Description |
|---|---|
| `WithRequestID()` | Generate request ID per request |
| `WithLogging(logger)` | Structured request logging |
| `WithRecovery(logger)` | Panic recovery |
| `WithCORS(cfg)` | CORS middleware |
| `WithShutdownTimeout(d)` | Graceful shutdown deadline |
| `WithTLS(...)` | TLS/HTTPS |

## Rules

- ✅ Use `http.NewServer(addr, handler, opts...)` — do not hand-roll `net/http.Server`.
- ✅ Compose middleware via options or `server.Use`.
- ✅ Pair with `govern/graceful` for shutdown.
- ✅ Keep handlers thin; delegate business logic to usecase layer.
- ❌ Do not bypass govern/http to wire `signal.Notify` shutdown manually.
- ❌ Do not block in handlers without respecting request context.

## Avoid

- Hand-rolled `net/http.Server` + `signal.Notify`.
- Ad-hoc shutdown ordering.
- Re-implementing middleware govern already provides (CORS, logging, recovery, request ID).

## Reference

Source: [`http/`](../../../../../../../http/).
