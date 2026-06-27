# JWT Authentication

JWT (JSON Web Token) authentication middleware and token management for Go HTTP servers.

## Features

- Token generation (access and refresh tokens)
- Token validation with comprehensive error handling
- HTTP middleware for route protection
- Custom claims with role-based authorization
- Configurable signing methods (HS256, HS384, HS512)
- Secure secret generation
- Path-based skipping for public routes
- Context-based user retrieval

## Installation

```bash
go get github.com/haipham22/govern/http/jwt
```

## Quick Start

### 1. Configuration

```go
import "github.com/haipham22/govern/http/jwt"

// Use defaults
config := jwt.DefaultConfig()
config.Secret = "your-secret-key-here"

// Or create custom config
config := &jwt.Config{
    Secret:          "your-secret-key",
    SigningMethod:   jwt.SigningMethodHS256,
    AccessDuration:  15 * time.Minute,
    RefreshDuration: 7 * 24 * time.Hour,
    Issuer:          "myapp",
}
```

### 2. Generate Tokens

```go
claims := &jwt.Claims{
    UserID:   "user123",
    Username: "johndoe",
    Email:    "john@example.com",
    Roles:    []string{"admin", "user"},
}

// Generate access token
accessToken, err := jwt.GenerateAccessToken(claims, config)

// Generate refresh token
refreshToken, err := jwt.GenerateRefreshToken(claims, config)
```

### 3. Setup Middleware

```go
import "net/http"

middlewareConfig := &jwt.MiddlewareConfig{
    Config:         config,
    TokenExtractor: jwt.DefaultTokenExtractor,
    ErrorHandler:   jwt.DefaultErrorHandler,
    SkipPaths:      []string{"/health", "/public", "/login"},
}

// Protect routes
mux := http.NewServeMux()
protectedHandler := jwt.Middleware(middlewareConfig)(mux)

http.ListenAndServe(":8080", protectedHandler)
```

### 4. Access User in Handlers

```go
func handleProfile(w http.ResponseWriter, r *http.Request) {
    // Get user from context
    claims, ok := jwt.GetCurrentUser(r)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    fmt.Fprintf(w, "Hello, %s!", claims.Username)

    // Or panic if user must be present (after JWT middleware)
    claims := jwt.MustGetCurrentUser(r)
}
```

## Token Generation

### Access Token

Short-lived token for API authentication:

```go
claims := &jwt.Claims{
    UserID:   "user123",
    Username: "johndoe",
    Email:    "john@example.com",
    Roles:    []string{"user"},
}

accessToken, err := jwt.GenerateAccessToken(claims, config)
```

### Refresh Token

Long-lived token for obtaining new access tokens:

```go
refreshToken, err := jwt.GenerateRefreshToken(claims, config)
```

### Secure Secret Generation

Generate cryptographically secure secrets:

```go
// Generate 32-byte secret (recommended)
secret, err := jwt.GenerateSecureSecret(32)

// Generate 64-byte secret (extra security)
secret, err := jwt.GenerateSecureSecret(64)
```

## Token Validation

### Validate Token

```go
claims, err := jwt.ValidateToken(tokenString, config)
if err != nil {
    // Handle error (expired, invalid signature, etc.)
    return err
}

fmt.Println("User ID:", claims.UserID)
fmt.Println("Roles:", claims.Roles)
```

### Refresh Access Token

```go
newAccessToken, err := jwt.RefreshAccessToken(refreshToken, config)
if err != nil {
    return err
}
```

## Claims Structure

```go
type Claims struct {
    UserID   string   `json:"user_id"`
    Username string   `json:"username"`
    Email    string   `json:"email,omitempty"`
    Roles    []string `json:"roles,omitempty"`
    jwt.RegisteredClaims  // Standard JWT claims (exp, iat, iss, etc.)
}
```

### Role-Based Authorization

```go
func handleAdmin(w http.ResponseWriter, r *http.Request) {
    claims := jwt.MustGetCurrentUser(r)

    // Check single role
    if !claims.HasRole("admin") {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // Check multiple roles (any match)
    if !claims.HasAnyRole("admin", "moderator") {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // Admin logic here
}
```

## Middleware Configuration

### Middleware Options

| Field           | Type                                | Default                | Description                     |
|-----------------|-------------------------------------|------------------------|---------------------------------|
| Config          | `*Config`                           | `DefaultConfig()`      | JWT configuration               |
| TokenExtractor  | `func(*http.Request) (string, err)` | DefaultTokenExtractor  | Extract token from request      |
| ErrorHandler    | `func(http.ResponseWriter, ...)`    | DefaultErrorHandler    | Handle authentication errors    |
| SkipPaths       | `[]string`                          | `[]`                   | Paths to skip authentication    |

### Custom Token Extractor

Extract token from custom location:

```go
customExtractor := func(r *http.Request) (string, error) {
    // Extract from query parameter
    token := r.URL.Query().Get("token")
    if token == "" {
        return "", jwt.ErrTokenMissing
    }
    return token, nil
}

middlewareConfig.TokenExtractor = customExtractor
```

### Custom Error Handler

Customize authentication error responses:

```go
customErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    // Return JSON errors
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{
        "error": err.Error(),
    })
}

middlewareConfig.ErrorHandler = customErrorHandler
```

### Skip Public Routes

```go
middlewareConfig.SkipPaths = []string{
    "/health",
    "/public",
    "/login",
    "/register",
    "/docs",
}
```

**Note**: Skip paths use prefix matching. `/api` skips `/api/users`, `/api/posts`, etc.

## Configuration Options

### Config Fields

| Field           | Type                | Default              | Description                      |
|-----------------|---------------------|----------------------|----------------------------------|
| Secret          | string              | **required**         | JWT signing secret key           |
| SigningMethod   | jwt.SigningMethod   | HS256                | JWT signing algorithm            |
| AccessDuration  | time.Duration       | 15 minutes           | Access token lifetime            |
| RefreshDuration | time.Duration       | 7 days               | Refresh token lifetime           |
| Issuer          | string              | "govern"             | JWT issuer claim                 |

### Signing Methods

Supported algorithms:

```go
jwt.SigningMethodHS256  // HMAC-SHA256 (recommended)
jwt.SigningMethodHS384  // HMAC-SHA384
jwt.SigningMethodHS512  // HMAC-SHA512
```

## Authentication Flow

### Login Flow

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    var creds struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&creds)

    // Validate credentials (check database, etc.)
    user, err := validateUser(creds.Username, creds.Password)
    if err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    // Create claims
    claims := &jwt.Claims{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        Roles:    user.Roles,
    }

    // Generate tokens
    accessToken, _ := jwt.GenerateAccessToken(claims, jwtConfig)
    refreshToken, _ := jwt.GenerateRefreshToken(claims, jwtConfig)

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    })
}
```

### Refresh Flow

```go
func handleRefresh(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    newAccessToken, err := jwt.RefreshAccessToken(req.RefreshToken, jwtConfig)
    if err != nil {
        http.Error(w, "invalid refresh token", http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "access_token": newAccessToken,
    })
}
```

## Error Handling

### Error Types

| Error                      | Code        | Description                        |
|----------------------------|-------------|------------------------------------|
| ErrSecretRequired          | invalid     | Secret key not provided            |
| ErrSigningMethodRequired   | invalid     | Signing method not provided        |
| ErrInvalidAccessDuration   | invalid     | Invalid access token duration      |
| ErrInvalidRefreshDuration  | invalid     | Invalid refresh token duration     |
| ErrTokenMissing            | unauthorized| Authorization header missing       |
| ErrTokenInvalid            | unauthorized| Invalid token format               |
| ErrTokenExpired            | unauthorized| Token has expired                  |
| ErrTokenMalformed          | unauthorized| Malformed token                    |
| ErrSignatureInvalid        | unauthorized| Invalid token signature            |
| ErrUnexpectedSigningMethod | unauthorized| Unexpected signing method          |

### Handling Errors

```go
import "github.com/haipham22/govern/errors"

_, err := jwt.ValidateToken(tokenString, config)
if errors.IsCode(err, errors.CodeUnauthorized) {
    // Authentication failed
    if err == jwt.ErrTokenExpired {
        // Token expired - prompt for refresh
    } else {
        // Other authentication error
    }
}
```

## Security Considerations

### Critical Security Practices

1. **Secret Key Storage**
   - Never hardcode secrets in source code
   - Use environment variables or secret management systems
   - Rotate secrets periodically
   - Use at least 32-byte secrets for HS256

   ```bash
   # Set via environment
   export JWT_SECRET="your-secret-key"
   ```

   ```go
   config.Secret = os.Getenv("JWT_SECRET")
   ```

2. **Token Duration**
   - Keep access tokens short-lived (5-15 minutes)
   - Keep refresh tokens longer-lived (7-30 days)
   - Implement refresh token rotation on critical systems

3. **HTTPS Only**
   - Always use HTTPS in production
   - Tokens sent over HTTP can be intercepted
   - Set secure cookie flags if storing tokens in cookies

4. **Token Storage (Client-side)**
   - Prefer httpOnly cookies for web apps
   - If using localStorage, implement XSS protection
   - Consider token binding to IP/session for sensitive operations

5. **Secret Generation**
   - Use `GenerateSecureSecret()` for production secrets
   - Never use weak or predictable secrets

6. **Claims Validation**
   - Always validate required claims (UserID)
   - Verify roles/permissions for protected resources
   - Implement token blacklisting for logout if needed

7. **Refresh Token Security**
   - Store refresh tokens securely (httpOnly cookies preferred)
   - Implement one-time use refresh tokens
   - Consider refresh token expiration and revocation

## API Reference

### Token Operations

```go
// Generate access token
func GenerateAccessToken(claims *Claims, config *Config) (string, error)

// Generate refresh token
func GenerateRefreshToken(claims *Claims, config *Config) (string, error)

// Validate token and return claims
func ValidateToken(tokenString string, config *Config) (*Claims, error)

// Refresh access token from refresh token
func RefreshAccessToken(refreshToken string, config *Config) (string, error)
```

### Middleware

```go
// Create JWT authentication middleware
func Middleware(config *MiddlewareConfig) func(http.Handler) http.Handler

// Get current user from request context (safe)
func GetCurrentUser(r *http.Request) (*Claims, bool)

// Get current user or panic (after JWT middleware)
func MustGetCurrentUser(r *http.Request) *Claims

// Extract token from Authorization header (Bearer token)
func DefaultTokenExtractor(r *http.Request) (string, error)

// Default error handler for middleware
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error)
```

### Configuration

```go
// Get default configuration
func DefaultConfig() *Config

// Validate configuration
func (c *Config) Validate() error

// Generate cryptographically secure secret
func GenerateSecureSecret(length int) (string, error)
```

### Claims Operations

```go
// Validate claims
func (c *Claims) Validate() error

// Check if user has specific role
func (c *Claims) HasRole(role string) bool

// Check if user has any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool
```

## Examples

### Complete Authentication Example

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/haipham22/govern/http/jwt"
)

var jwtConfig = &jwt.Config{
    Secret:          "your-secret-key",  // Load from env in production
    SigningMethod:   jwt.SigningMethodHS256,
    AccessDuration:  15 * time.Minute,
    RefreshDuration: 7 * 24 * time.Hour,
    Issuer:          "myapp",
}

func main() {
    mux := http.NewServeMux()

    // Public routes
    mux.HandleFunc("/login", handleLogin)
    mux.HandleFunc("/refresh", handleRefresh)

    // Protected routes
    mux.HandleFunc("/profile", jwt.Middleware(&jwt.MiddlewareConfig{
        Config:         jwtConfig,
        TokenExtractor: jwt.DefaultTokenExtractor,
        ErrorHandler:   jwt.DefaultErrorHandler,
        SkipPaths:      []string{"/login", "/refresh"},
    })(http.HandlerFunc(handleProfile)).ServeHTTP

    http.ListenAndServe(":8080", mux)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Validate user credentials
    claims := &jwt.Claims{
        UserID:   "user123",
        Username: "johndoe",
        Roles:    []string{"user"},
    }

    accessToken, _ := jwt.GenerateAccessToken(claims, jwtConfig)
    refreshToken, _ := jwt.GenerateRefreshToken(claims, jwtConfig)

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    })
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    newAccessToken, err := jwt.RefreshAccessToken(req.RefreshToken, jwtConfig)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "access_token": newAccessToken,
    })
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
    claims := jwt.MustGetCurrentUser(r)
    json.NewEncoder(w).Encode(claims)
}
```

## Testing

```bash
go test -race ./http/jwt/...
```

## License

TBD
