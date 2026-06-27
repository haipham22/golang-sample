# Echo Utilities

The `echo` package provides utilities for using [Echo](https://echo.labstack.com/) framework with Govern, including JWT authentication middleware, Swagger UI integration, and handler wrapping utilities.

## Features

- **JWT Authentication**: Echo-specific JWT middleware using `http/jwt`
- **Swagger UI Integration**: Configuration utilities for Swagger UI with authentication support
- **Handler Wrapping**: Reuse standard `http.Handler` with Echo via `WrapHandler()`
- **Context Helpers**: Utilities for retrieving current user from Echo context
- **TrimStrings Middleware**: Automatic whitespace trimming from JSON request bodies

## Installation

```bash
go get github.com/haipham22/govern/http/echo
```

## JWT Authentication

Create JWT authentication middleware for Echo routes:

```go
package main

import (
    httpEcho "github.com/haipham22/govern/http/echo"
    "github.com/haipham22/govern/http/jwt"
    "github.com/labstack/echo/v4"
    "net/http"
)

func main() {
    e := echo.New()

    // JWT configuration
    jwtConfig := &jwt.MiddlewareConfig{
        Config:         jwt.DefaultConfig(),
        TokenExtractor: jwt.DefaultTokenExtractor,
        SkipPaths:      []string{"/health", "/login"},
    }
    jwtConfig.Config.Secret = "your-secret-key"

    // Public routes
    e.GET("/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
    })

    // Protected routes group
    api := e.Group("/api", httpEcho.JWTMiddleware(jwtConfig))
    api.GET("/profile", func(c echo.Context) error {
        claims := httpEcho.MustGetCurrentUser(c)
        return c.JSON(http.StatusOK, map[string]interface{}{
            "user_id":  claims.UserID,
            "username": claims.Username,
        })
    })

    e.Start(":8080")
}
```

### Context Helpers

```go
func handler(c echo.Context) error {
    // Returns claims and true if found, nil and false otherwise
    claims, ok := httpEcho.GetCurrentUser(c)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
    }

    // Panics if user not found (use after JWT middleware)
    claims := httpEcho.MustGetCurrentUser(c)

    // Get individual fields
    userID, ok := httpEcho.GetUserID(c)
    username, ok := httpEcho.GetUsername(c)

    return c.JSON(http.StatusOK, claims)
}
```

## Swagger UI Integration

Configure Swagger UI for your Echo application using `WithEchoSwagger`:

### Quick Start

```go
import httpEcho "github.com/haipham22/govern/http/echo"

e := echo.New()

// Setup Swagger UI
httpEcho.WithEchoSwagger(e,
    httpEcho.WithSwaggerEnabled(true),
    httpEcho.WithSwaggerInfo(&httpEcho.SwaggerInfo{
        Title:       "My API",
        Description: "Sample API server",
        Version:     "1.0",
    }),
)

// Add your routes
e.GET("/users", getUsers)

e.Start(":8080")
```

Access Swagger UI at: `http://localhost:8080/swagger/`

### Configuration Options

```go
httpEcho.WithEchoSwagger(e,
    // Enable/disable Swagger UI
    httpEcho.WithSwaggerEnabled(true),

    // Custom path (default: /swagger/*)
    httpEcho.WithSwaggerPath("/api/docs/*"),

    // API metadata
    httpEcho.WithSwaggerInfo(&httpEcho.SwaggerInfo{
        Title:          "User Management API",
        Description:    "API for managing users",
        Version:        "2.0",
        Host:           "api.example.com",
        BasePath:       "/v1",
        TermsOfService: "https://example.com/terms",
        ContactName:    "API Support",
        ContactEmail:   "support@example.com",
        LicenseName:    "MIT",
        LicenseURL:     "https://opensource.org/licenses/MIT",
    }),

    // Authentication
    httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
        Type:        "Bearer",
        Description: "JWT token",
        Name:        "Authorization",
        In:          "header",
    }),
)
```

### Authentication Types

**Bearer Token (JWT):**
```go
httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
    Type:        "Bearer",
    Description: "JWT token for authentication",
    Name:        "Authorization",
    In:          "header",
})
```

**API Key:**
```go
httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
    Type:        "ApiKey",
    Description: "API key in header",
    Name:        "X-API-Key",
    In:          "header",
})
```

**Basic Auth:**
```go
httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
    Type:        "Basic",
    Description: "Basic authentication",
})
```

**OAuth2:**
```go
httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
    Type:            "OAuth2",
    Description:     "OAuth2 flow",
    Flow:            "implicit",
    AuthorizationURL: "https://example.com/oauth/authorize",
    TokenURL:        "https://example.com/oauth/token",
    Scopes: map[string]string{
        "read":  "Read access",
        "write": "Write access",
    },
})
```

### With Swag Tool Integration

For full Swagger documentation with swag:

1. Install swag:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Add annotations to handlers:

```go
// @Summary Get user by ID
// @Description Retrieve user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [get]
// @Security Bearer
func getUser(c echo.Context) error {
    // Handler logic
}
```

3. Generate docs:

```bash
swag init -g cmd/api/main.go
```

4. Import docs package and setup Swagger:

```go
import _ "myapi/docs"  // Generated docs

httpEcho.WithEchoSwagger(e,
    httpEcho.WithSwaggerEnabled(true),
    httpEcho.WithSwaggerAuth(&httpEcho.SwaggerAuth{
        Type: "Bearer",
        Description: "JWT token",
        Name: "Authorization",
        In: "header",
    }),
)
```

### Security

**Swagger UI is disabled by default for security.**

Only enable in development or behind authentication:

```go
enableSwagger := os.Getenv("APP_ENV") == "development"

httpEcho.WithEchoSwagger(e,
    httpEcho.WithSwaggerEnabled(enableSwagger),
    httpEcho.WithSwaggerPath("/swagger/*"),
)
```

**Production recommendations:**
- Keep Swagger UI disabled in production
- Or protect it with authentication middleware
- Use environment variables for configuration
- Never expose Swagger UI on public internet without auth

## Handler Wrapping

Reuse standard `http.Handler` with Echo:

```go
import (
    "net/http"
    httpEcho "github.com/haipham22/govern/http/echo"
)

// Reuse standard http.Handler with Echo
var myHandler http.Handler = // ...

e.GET("/legacy", httpEcho.WrapHandler(myHandler))
```

## TrimStrings Middleware

Automatically trim whitespace from string fields in JSON request bodies:

```go
import httpEcho "github.com/haipham22/govern/http/echo"

e := echo.New()

// Apply to individual routes
e.POST("/api/register", handleRegister, httpEcho.TrimStrings)

// Apply to route groups
api := e.Group("/api")
api.Use(httpEcho.TrimStrings)
api.POST("/users", createUser)

func handleRegister(c echo.Context) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.Bind(&req); err != nil {
        return err
    }

    // All string fields are already trimmed!
    // req.Username, req.Email, req.Password have no leading/trailing whitespace
    return c.JSON(http.StatusOK, req)
}
```

**Features:**
- Recursively trims all string values in JSON request bodies
- Handles nested objects and arrays
- Uses sonic for high-performance JSON parsing (~2x faster than encoding/json)
- Graceful fallback for invalid JSON (passes through unchanged)
- Fast-path optimization for strings without whitespace
- Zero overhead for non-JSON requests

**Example:**

```go
// Request body:
// {
//   "username": "  john_doe  ",
//   "email": "  john@example.com  ",
//   "profile": {
//     "name": "  John Doe  ",
//     "bio": "  Software Engineer  "
//   },
//   "tags": ["  go  ", "  echo  ", "  api  "]
// }

// After TrimStrings middleware:
// {
//   "username": "john_doe",
//   "email": "john@example.com",
//   "profile": {
//     "name": "John Doe",
//     "bio": "Software Engineer"
//   },
//   "tags": ["go", "echo", "api"]
// }
```

## Middleware Compatibility

### IMPORTANT: Do NOT Mix Middleware Types

Echo middleware (`echo.MiddlewareFunc`) and Govern middleware (`http.Middleware`) are **not compatible** due to:

1. Different error handling flows (Echo returns errors, http middleware doesn't)
2. Different context types (`echo.Context` vs `http.Request`)
3. Response writer differences (`echo.Response` vs `http.ResponseWriter`)
4. Complex conversion introduces bugs and performance overhead

**Instead:**
- Use Echo middleware for Echo routes
- Use Govern middleware for `http.Handler` routes
- Use `WrapHandler()` to reuse `http.Handler` with Echo if needed

### Using Echo Middleware

```go
import (
    "github.com/labstack/echo/v4/middleware"
)

e := echo.New()

// Use Echo's built-in middleware
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORS())
```

## Testing

All tests run with race detector:

```bash
go test -race ./http/echo/...
```

## Best Practices

1. **Use Echo middleware** for Echo routes (logging, recovery, CORS, TrimStrings)
2. **Use WrapHandler()** to reuse standard `http.Handler` with Echo
3. **Apply TrimStrings early** in the middleware chain (before validation/handlers)
4. **Always pass context** as first parameter to downstream operations
5. **Test with race detector**: `go test -race ./http/echo/...`
6. **Run golangci-lint**: `golangci-lint run ./http/echo/...`

## Middleware Order

Recommended middleware order for Echo:

```go
e := echo.New()

// 1. Request ID (outermost)
e.Use(middleware.RequestID())

// 2. Logger
e.Use(middleware.Logger())

// 3. Recovery
e.Use(middleware.Recover())

// 4. CORS
e.Use(middleware.CORS())

// 5. TrimStrings (before validation/handlers)
e.Use(httpEcho.TrimStrings)

// 6. JWT authentication (for protected routes)
api := e.Group("/api", httpEcho.JWTMiddleware(jwtConfig))

// 7. Your handlers
api.GET("/users", getUsers)
```

## Examples

See `swagger_auth_example.go` for Swagger authentication examples.
