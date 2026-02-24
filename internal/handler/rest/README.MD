# REST Handlers Package

This package contains HTTP handlers for the REST API.

## Structure

```
internal/handler/rest/
├── controllers/
│   ├── auth/       # Authentication controllers
│   │   ├── auth.go       # Main auth controller with login/register
│   │   └── converter.go  # Request/response converters
│   └── health/     # Health check controllers
│       └── handler.go    # Health check handler
├── middlewares/    # HTTP middlewares
│   ├── compression.go    # Gzip compression middleware
│   ├── cors.go           # CORS middleware
│   ├── ratelimit.go      # Rate limiting middleware
│   └── security.go       # Security headers middleware
├── swagger/        # Swagger documentation
├── routes.go       # Route registration
├── handler.go      # Main handler setup
└── wire.go         # Wire dependency injection
```

## Adding New Handlers

1. Create handler file in appropriate subdirectory
2. Define handler struct with dependencies
3. Implement handler methods
4. Register routes in `routes.go`
5. Add Wire provider in `wire.go`
6. Write tests

## Example

```go
package myhandler

import "github.com/labstack/echo/v4"

type Handler struct {
	// dependencies
}

func (h *Handler) GetItem(c echo.Context) error {
	// handler logic
}
```
