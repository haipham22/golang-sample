# govern/http/middleware

Import: `github.com/haipham22/govern/http/middleware`

Common HTTP middleware: CORS, logging, recovery, request ID, compression, security headers.

## Use When

- Need CORS, structured logging, panic recovery, request ID, or security headers.

## CORS Example

```go
import "github.com/haipham22/govern/http/middleware"

server := http.NewServer(":8080", handler,
    http.WithCORS(&middleware.CORSConfig{
        AllowedOrigins:   []string{"https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    }),
)
```

## Rules

- ✅ Apply middleware via `http.NewServer` options or `server.Use`.
- ✅ Prefer govern middleware over separate third-party libs.
- ✅ Set explicit `AllowedOrigins` in production — avoid `*` with credentials.
- ✅ Always run recovery + request ID near the top of the chain.
- ❌ Do not stack duplicate middleware (e.g. two logging middlewares).
- ❌ Do not allow credential cookies with wildcard CORS origin.

## Avoid

- Pulling separate CORS/security libs for concerns govern covers.
- Inconsistent middleware ordering across services.

## Reference

Source: [`http/middleware/`](../../../../../../../http/middleware/).
