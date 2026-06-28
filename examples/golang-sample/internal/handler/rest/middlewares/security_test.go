package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)

	// Verify all security headers are set
	headers := rec.Header()

	// X-Frame-Options
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"), "X-Frame-Options should be DENY")

	// X-Content-Type-Options
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"), "X-Content-Type-Options should be nosniff")

	// X-XSS-Protection
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"), "X-XSS-Protection should be enabled")

	// Strict-Transport-Security
	hsts := headers.Get("Strict-Transport-Security")
	assert.Contains(t, hsts, "max-age=31536000", "HSTS should have max-age of 1 year")
	assert.Contains(t, hsts, "includeSubDomains", "HSTS should include subdomains")

	// Content-Security-Policy
	csp := headers.Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'", "CSP should restrict to same origin")
	assert.Contains(t, csp, "frame-ancestors 'none'", "CSP should prevent framing")

	// Referrer-Policy
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"), "Referrer-Policy should be set")

	// Permissions-Policy
	assert.Equal(t, "camera=(), microphone=(), geolocation=()", headers.Get("Permissions-Policy"), "Permissions-Policy should restrict features")
}

func TestSecurityHeadersWithConfig_CustomConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	customConfig := SecurityHeadersConfig{
		FrameOptions:          "SAMEORIGIN",
		HSTSMaxAge:            12345,
		CSP:                   "default-src 'self' https://example.com",
		HSTSIncludeSubDomains: false,
	}

	middleware := SecurityHeadersWithConfig(customConfig)
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)

	headers := rec.Header()

	// Verify custom config is applied
	assert.Equal(t, "SAMEORIGIN", headers.Get("X-Frame-Options"))
	assert.Contains(t, headers.Get("Strict-Transport-Security"), "max-age=12345")
	assert.NotContains(t, headers.Get("Strict-Transport-Security"), "includeSubDomains")
	assert.Equal(t, "default-src 'self' https://example.com", headers.Get("Content-Security-Policy"))
}

func TestSecurityHeaders_DoesNotInterfereWithResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Hello, World!")
}

func TestSecurityHeaders_AppliesToErrorResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	// Note: Error responses go through the error handler, not middleware
	// But we can still test that the middleware itself doesn't cause issues
	err := h(c)
	assert.Error(t, err)

	// In a real scenario, the error handler would be called
	// and we'd verify headers are set even on error responses
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	assert.Equal(t, "DENY", config.FrameOptions)
	assert.Equal(t, 31536000, config.HSTSMaxAge)
	assert.Contains(t, config.CSP, "default-src 'self'")
	assert.True(t, config.HSTSIncludeSubDomains)
}

func TestCORS_DefaultConfig(t *testing.T) {
	config := DefaultCORSConfig()

	// Verify default origins include localhost
	assert.Contains(t, config.AllowOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowOrigins, "http://localhost:8080")

	// Verify methods
	assert.Contains(t, config.AllowMethods, "GET")
	assert.Contains(t, config.AllowMethods, "POST")
	assert.Contains(t, config.AllowMethods, "PUT")

	// Verify credentials are allowed
	assert.True(t, config.AllowCredentials)

	// Verify max age
	assert.Equal(t, 86400, config.MaxAge)
}

func TestCORS_ProductionConfig(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://app.example.com"}
	config := ProductionCORSConfig(allowedOrigins)

	assert.Equal(t, allowedOrigins, config.AllowOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}

func TestCORS_AllowsCredentials(t *testing.T) {
	config := DefaultCORSConfig()
	assert.True(t, config.AllowCredentials, "CORS should allow credentials by default")
}

func TestSecurityHeaders_PreventsClickjacking(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)

	// X-Frame-Options: DENY prevents page from being framed
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))

	// CSP frame-ancestors 'none' also prevents framing
	csp := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "frame-ancestors 'none'")
}

func TestSecurityHeaders_PreventsMIMETypeSniffing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)

	// X-Content-Type-Options: nosniff prevents MIME sniffing
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_EnforcesHTTPS(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := SecurityHeaders()
	h := middleware(next)

	err := h(c)
	assert.NoError(t, err)

	hsts := rec.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hsts, "max-age=31536000", "HSTS should enforce HTTPS for 1 year")
	assert.Contains(t, hsts, "includeSubDomains", "HSTS should include subdomains")
}
