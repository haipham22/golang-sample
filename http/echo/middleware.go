package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// WrapHandler adapts http.Handler to Echo handler func.
// Use this to reuse standard http handlers with Echo.
func WrapHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

// Middleware Usage Guidelines:
//
// DO NOT convert between Echo and Govern middleware types due to:
// 1. Different error handling flows (Echo returns errors, http middleware doesn't)
// 2. Different context types (echo.Context vs http.Request)
// 3. Response writer differences (echo.Response vs http.ResponseWriter)
// 4. Complex conversion introduces bugs and performance overhead
//
// Instead:
// - Use Echo middleware (echo.MiddlewareFunc) for Echo routes
// - Use Govern middleware (http.Middleware) for http.Handler routes
// - Use WrapHandler() to reuse http.Handler with Echo if needed
