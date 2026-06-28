package middlewares

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// BodyLimit returns a middleware that limits the size of request body
// Default: 1MB max for security
func BodyLimit() echo.MiddlewareFunc {
	return BodyLimitWithConfig("1M")
}

// BodyLimitWithConfig returns a middleware with custom body size limit
func BodyLimitWithConfig(limit string) echo.MiddlewareFunc {
	return echomiddleware.BodyLimit(limit)
}
