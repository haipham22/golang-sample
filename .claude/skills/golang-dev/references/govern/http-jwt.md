# govern/http/jwt

Import: `github.com/haipham22/govern/http/jwt`

JWT authentication middleware. Validates tokens, populates request context with claims.

## Use When

- App needs JWT auth on protected routes.

## Typical Wiring

```go
import (
    "github.com/haipham22/govern/http/jwt"
)

protected := router.Group("/api")
protected.Use(jwt.Middleware(secret))
```

## Rules

- ✅ Load JWT secret from config/env, never hardcode.
- ✅ Reject missing/invalid tokens with consistent error mapping.
- ✅ Map JWT middleware errors to app error envelope at boundary.
- ❌ Do not roll custom JWT validation.
- ❌ Do not scatter token parsing across handlers.
- ❌ Do not log token payloads (PII / credentials).

## Security

- Use constant-time comparison where applicable.
- Short-lived access tokens; rotate secrets.
- Validate `exp`, `iat`, `nbf`, issuer/audience.

## Avoid

- Custom JWT parsing duplicated per handler.
- Storing raw tokens in logs.

## Reference

Source: [`http/jwt/`](../../../../../../../http/jwt/). Uses `github.com/golang-jwt/jwt/v5`.
