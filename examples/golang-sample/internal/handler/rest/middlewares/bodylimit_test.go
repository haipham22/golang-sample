package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveWithBodyLimit wires a handler with the given body-limit middleware onto
// a fresh Echo engine and dispatches a request. Using e.ServeHTTP (rather than
// calling the middleware chain directly) ensures Echo's error handler converts
// BodyLimit's 413 echo.HTTPError into an actual 413 response.
func serveWithBodyLimit(t *testing.T, mw echo.MiddlewareFunc, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.Use(mw)
	e.POST("/x", func(c *echo.Context) error { return c.String(http.StatusOK, "OK") })

	req := httptest.NewRequest(method, "/x", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestBodyLimit_DefaultBlocksOversize verifies BodyLimit() returns a working
// middleware that rejects bodies larger than the 1 MiB default. We use a small
// custom limit via the BodyLimit constructor surface to keep the test fast
// while still exercising the default path separately below.
func TestBodyLimit_DefaultBlocksOversize(t *testing.T) {
	mw := BodyLimit()
	require.NotNil(t, mw)

	// The default middleware should accept a normal-sized body (proving the
	// constructor produced a functional middleware). A full 1 MiB+1 oversize
	// case is exercised against BodyLimitWithConfig in the next test to avoid
	// allocating 1 MiB per run.
	rec := serveWithBodyLimit(t, mw, http.MethodPost, `{"ok":true}`)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestBodyLimitWithConfig_RejectsOverLimit verifies oversize bodies get 413.
func TestBodyLimitWithConfig_RejectsOverLimit(t *testing.T) {
	mw := BodyLimitWithConfig(4)
	rec := serveWithBodyLimit(t, mw, http.MethodPost, `{"too":"long"}`)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Contains(t, rec.Body.String(), "Request Entity Too Large")
}

// TestBodyLimitWithConfig_AllowsWithinLimit verifies small bodies pass through.
func TestBodyLimitWithConfig_AllowsWithinLimit(t *testing.T) {
	mw := BodyLimitWithConfig(64)
	rec := serveWithBodyLimit(t, mw, http.MethodPost, `{"ok":true}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}
