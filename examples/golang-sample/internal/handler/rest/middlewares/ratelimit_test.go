package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_AllowsRequestsUnderLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	next := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := RateLimitWithConfig(context.Background(), RateLimiterConfig{
		RequestsPerMinute: 5,
		WindowSize:        60,
	})

	h := middleware(next)

	// Make 5 requests (should all succeed)
	for i := 0; i < 5; i++ {
		handlerCalled = false
		req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.True(t, handlerCalled, "Handler should be called for request %d", i+1)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimit_BlocksRequestsOverLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RateLimitWithConfig(context.Background(), RateLimiterConfig{
		RequestsPerMinute: 3,
		WindowSize:        60,
	})

	h := middleware(next)

	// Set a consistent IP for testing
	testIP := "192.168.1.1:1234"

	// Make 3 requests (should succeed)
	for i := 0; i < 3; i++ {
		req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
		req.RemoteAddr = testIP
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
	}

	// 4th request should be rate limited
	req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = testIP
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "Request should be rate limited")
	assert.Contains(t, rec.Body.String(), "Too many requests", "Should return rate limit error message")
}

func TestRateLimit_SlidingWindow(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	// Short window for testing
	middleware := RateLimitWithConfig(context.Background(), RateLimiterConfig{
		RequestsPerMinute: 2,
		WindowSize:        2, // 2 seconds
	})

	h := middleware(next)

	testIP := "192.168.1.2:1234"

	// Make 2 requests (should succeed)
	for i := 0; i < 2; i++ {
		req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
		req.RemoteAddr = testIP
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 3rd request should be blocked
	req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = testIP
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// After window expires, request should succeed again
	req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = testIP
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code, "Request should succeed after window expires")
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RateLimitWithConfig(context.Background(), RateLimiterConfig{
		RequestsPerMinute: 2,
		WindowSize:        60,
	})

	h := middleware(next)

	// Make 2 requests from IP1 (should succeed)
	for i := 0; i < 2; i++ {
		req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 3rd request from IP1 should be blocked
	req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// But IP2 should still be able to make requests
	for i := 0; i < 2; i++ {
		req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
		req.RemoteAddr = "192.168.1.2:5678"
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code, "IP2 request %d should succeed", i+1)
	}
}

func TestRateLimit_WithRealIP(t *testing.T) {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()

	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RateLimitWithConfig(context.Background(), RateLimiterConfig{
		RequestsPerMinute: 1,
		WindowSize:        60,
	})

	h := middleware(next)

	// First request should succeed
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Second request should be rate limited
	req = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 10, config.RequestsPerMinute, "Default should allow 10 requests per minute")
	assert.Equal(t, 60, config.WindowSize, "Default window should be 60 seconds")
}

func TestRateLimit_ContextCancellation(t *testing.T) {
	// Test that goroutine cleanup works when context is cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create middleware with cancellable context
	middleware := RateLimitWithConfig(ctx, DefaultRateLimiterConfig())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	h := middleware(next)

	// Make a request to ensure limiter is created
	err := h(c)
	assert.NoError(t, err)

	// Cancel context to trigger goroutine cleanup
	cancel()

	// Give goroutine time to exit
	time.Sleep(100 * time.Millisecond)

	// If we reach here without deadlock/panic, cleanup worked
	// Note: In real scenario, goroutine exits cleanly
}
