package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCORS invokes the CORS middleware against a configured Echo engine and
// returns the response recorder. Used to exercise CORS()/CORSWithConfig() so
// those constructors are covered end-to-end (preflight + actual request).
func runCORS(t *testing.T, mw echo.MiddlewareFunc, method, origin, reqOrigin string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.Use(mw)
	e.GET("/x", func(c *echo.Context) error { return c.String(http.StatusOK, "OK") })

	req := httptest.NewRequest(method, "/x", nil)
	req.Header.Set(echo.HeaderOrigin, reqOrigin)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestCORS_DefaultsAllowsLocalhost3000 verifies CORS() returns a middleware
// that echoes CORS headers for an allowed origin.
func TestCORS_DefaultsAllowsLocalhost3000(t *testing.T) {
	mw := CORS()
	require.NotNil(t, mw)

	rec := runCORS(t, mw, http.MethodGet, "", "http://localhost:3000")
	assert.NotEqual(t, http.StatusForbidden, rec.Code)
	// Echo's CORS sets Access-Control-Allow-Origin on allowed cross-origin.
	assert.Equal(t, "http://localhost:3000", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}

// TestCORS_PreflightReturnsOK verifies CORS() handles OPTIONS preflight with
// the default config (allowed methods/headers).
func TestCORS_PreflightReturnsOK(t *testing.T) {
	mw := CORS()

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set(echo.HeaderOrigin, "http://localhost:8080")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, "Content-Type,Authorization")
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Use(mw)
	e.GET("/x", func(c *echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:8080", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))
}

// TestCORSWithConfig_CustomOrigins verifies a custom config is applied.
func TestCORSWithConfig_CustomOrigins(t *testing.T) {
	mw := CORSWithConfig(ProductionCORSConfig([]string{"https://prod.example.com"}))

	rec := runCORS(t, mw, http.MethodGet, "", "https://prod.example.com")
	assert.Equal(t, "https://prod.example.com", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}

// TestCORSWithConfig_DisallowedOriginOmitsHeader verifies the middleware does
// NOT add Allow-Origin when the origin is not in the allowlist.
func TestCORSWithConfig_DisallowedOriginOmitsHeader(t *testing.T) {
	mw := CORSWithConfig(ProductionCORSConfig([]string{"https://prod.example.com"}))

	rec := runCORS(t, mw, http.MethodGet, "", "https://evil.example.com")
	assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}
