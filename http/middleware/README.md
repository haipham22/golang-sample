# middleware

Common HTTP middleware for logging, recovery, CORS, and request ID tracking.

## Features

- **Logging**: Request/response logging with timing, status codes, and bytes written
- **Recovery**: Panic recovery with stack trace logging
- **Request ID**: Unique request identification via headers and context
- **CORS**: Configurable Cross-Origin Resource Sharing with preflight support
- **TrimStrings**: Automatic whitespace trimming from JSON string fields
- **Composable**: Chain multiple middleware easily
- **Production-ready**: Battle-tested with comprehensive test coverage

## Installation

```bash
go get github.com/haipham22/govern/http/middleware
```

## Middleware Types

### Logging

Logs incoming requests and completed responses with timing and metadata.

```go
import (
    "github.com/haipham22/govern/http/middleware"
    "github.com/haipham22/govern/log"
)

logger := log.Default()
h := middleware.Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}))
```

**Logged Fields:**
- Request ID
- Method, path, query
- Remote address, user agent
- Response status, duration
- Bytes written

### Recovery

Recovers from panics, logs stack trace, returns 500 error.

```go
logger := log.Default()
h := middleware.Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    panic("something went wrong") // Recovered
}))
```

**Logged Fields:**
- Error value
- Stack trace
- Request ID (if available)
- Method and path

### Request ID

Adds unique request ID to context and response header.

```go
h := middleware.RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Get request ID from context
    reqID := r.Context().Value(middleware.RequestIDKeyVal)
    w.Write([]byte("Request ID: " + reqID.(string)))
}))
```

**Behavior:**
- Uses `X-Request-ID` header if present
- Generates UUID v4 if missing
- Sets ID in response header and request context

### CORS

Configurable CORS middleware with preflight request support.

```go
config := &middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           86400,
}
h := middleware.CORS(config)(http.HandlerFunc(handler))
```

**Default Configuration:**

```go
config := middleware.DefaultCORSConfig()
// AllowedOrigins:   ["*"]
// AllowedMethods:   ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
// AllowedHeaders:   ["Origin", "Content-Type", "Authorization"]
// AllowCredentials: false
// MaxAge:           86400 (24 hours)
```

### TrimStrings (Echo)

Automatically trims whitespace from string fields in JSON request bodies.

```go
import "github.com/haipham22/govern/http/middleware"

e := echo.New()

// Apply TrimStrings middleware to a route or group
e.POST("/api/register",
    middleware.TrimStrings(handleRegister),
)
```

**Behavior:**
- Recursively trims all string values in JSON request bodies
- Handles nested objects and arrays
- Uses sonic for high-performance JSON parsing
- Graceful fallback for invalid JSON (passes through unchanged)
- Fast-path optimization for strings without whitespace

**Example:**

```go
// Request body: {"username": "  john  ", "email": "  john@example.com  "}
// After middleware: {"username": "john", "email": "john@example.com"}

func handleRegister(c echo.Context) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    if err := c.Bind(&req); err != nil {
        return err
    }
    // req.Username and req.Email are already trimmed
    return c.JSON(http.StatusOK, req)
}
```

**Performance:**
- Uses [sonic](https://github.com/bytedance/sonic) for fast JSON operations
- Benchmarks show ~2x faster than standard encoding/json
- Minimal overhead for requests without whitespace

## Middleware Chaining

Chain middleware by composing them:

```go
logger := log.Default()

handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})

// Chain: RequestID -> Logging -> Recovery -> Handler
wrapped := middleware.RequestID()(
    middleware.Logging(logger)(
        middleware.Recovery(logger)(handler),
    ),
)
```

**Recommended Order (net/http):**
1. RequestID (outermost)
2. Logging
3. Recovery
4. Custom middleware
5. Handler (innermost)

**Recommended Order (Echo):**
1. RequestID (outermost)
2. Logging
3. Recovery
4. TrimStrings (before validation/handlers)
5. Custom middleware
6. Handler (innermost)

## Configuration Options

### CORSConfig

| Field             | Type      | Description                    | Default                    |
|-------------------|-----------|--------------------------------|----------------------------|
| AllowedOrigins    | []string  | Permitted origins              | ["*"]                      |
| AllowedMethods    | []string  | Permitted HTTP methods         | GET, POST, PUT, DELETE, OPTIONS |
| AllowedHeaders    | []string  | Permitted request headers      | Origin, Content-Type, Authorization |
| ExposedHeaders    | []string  | Exposed response headers       | -                          |
| AllowCredentials  | bool      | Allow credentials flag         | false                      |
| MaxAge            | int       | Preflight cache age (seconds)  | 86400 (24 hours)           |

## API Reference

### Middleware Functions

| Function                      | Signature                           | Description                         |
|-------------------------------|-------------------------------------|-------------------------------------|
| `Logging`                     | `(*zap.SugaredLogger) Middleware`   | Request/response logger            |
| `Recovery`                    | `(*zap.SugaredLogger) Middleware`   | Panic recovery                     |
| `RequestID`                   | `() Middleware`                     | Request ID injection                |
| `CORS`                        | `(*CORSConfig) Middleware`          | CORS handling                      |
| `DefaultCORSConfig`           | `() *CORSConfig`                    | Default CORS configuration         |
| `TrimStrings`                 | `(echo.HandlerFunc) echo.HandlerFunc` | Trim JSON string whitespace (Echo) |

### Types

| Type         | Description                       |
|--------------|-----------------------------------|
| `Middleware` | `func(http.Handler) http.Handler` |
| `CORSConfig` | CORS configuration struct         |
| `RequestIDKey` | Context key type for request ID |

### Context Keys

```go
const RequestIDKeyVal RequestIDKey = "request_id"
```

Access request ID in handlers:

```go
reqID := r.Context().Value(middleware.RequestIDKeyVal).(string)
```

## Complete Example

```go
package main

import (
    "net/http"

    "github.com/haipham22/govern/http/middleware"
    "github.com/haipham22/govern/log"
)

func main() {
    logger := log.Default()

    // Your handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := r.Context().Value(middleware.RequestIDKeyVal).(string)
        w.Write([]byte("Hello! Request ID: " + reqID))
    })

    // Chain middleware
    wrapped := middleware.RequestID()(
        middleware.Logging(logger)(
            middleware.Recovery(logger)(
                middleware.CORS(middleware.DefaultCORSConfig())(
                    handler,
                ),
            ),
        ),
    )

    http.ListenAndServe(":8080", wrapped)
}
```

## Testing

```go
func TestLogging(t *testing.T) {
    logger, buf := createTestLogger(t)
    mw := middleware.Logging(logger)

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("response"))
    })

    rec := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/test", http.NoBody)

    wrapped := mw(handler)
    wrapped.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
    // Parse logs from buf for assertions
}
```

## Best Practices

1. **Always use Recovery middleware** in production to prevent server crashes
2. **Place RequestID outermost** so all middleware have access to request ID
3. **Use Logging middleware** to track request flow and debug issues
4. **Configure CORS carefully** - avoid wildcard origins in production
5. **Test middleware chains** to ensure proper order and behavior
6. **Handle nil logger** - ensure logger is initialized before use

## Thread Safety

All middleware functions are thread-safe and can be used concurrently.
