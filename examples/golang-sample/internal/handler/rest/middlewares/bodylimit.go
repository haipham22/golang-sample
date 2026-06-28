package middlewares

import (
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

// bodyLimit1MB mirrors the previous v4 "1M" body limit (1 MiB in bytes).
const bodyLimit1MB = 1 << 20

// BodyLimit returns a middleware that limits the size of request body
// Default: 1MB max for security
func BodyLimit() echo.MiddlewareFunc {
	return BodyLimitWithConfig(bodyLimit1MB)
}

// BodyLimitWithConfig returns a middleware with custom body size limit (bytes).
// Echo v5's BodyLimit takes an int64 byte count rather than the v4 "1M"-style string.
func BodyLimitWithConfig(limit int64) echo.MiddlewareFunc {
	return echomiddleware.BodyLimit(limit)
}
