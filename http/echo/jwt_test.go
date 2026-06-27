package echo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpEcho "github.com/haipham22/govern/http/echo"
	"github.com/haipham22/govern/http/jwt"
)

func TestJWTMiddleware(t *testing.T) {
	config := &jwt.MiddlewareConfig{
		Config: jwt.DefaultConfig(),
	}
	config.Secret = "test-secret"
	config.TokenExtractor = jwt.DefaultTokenExtractor
	config.ErrorHandler = jwt.DefaultErrorHandler
	config.SkipPaths = []string{"/health"}

	e := echo.New()
	e.Use(httpEcho.JWTMiddleware(config))

	e.GET("/protected", func(c echo.Context) error {
		claims, ok := httpEcho.GetCurrentUser(c)
		require.True(t, ok)
		return c.JSON(http.StatusOK, claims)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test protected route without token
	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Test with valid token
	claims := &jwt.Claims{
		UserID:   "user123",
		Username: "testuser",
	}
	token, err := jwt.GenerateAccessToken(claims, config.Config)
	require.NoError(t, err)

	req = httptest.NewRequest("GET", "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test skip path
	req = httptest.NewRequest("GET", "/health", http.NoBody)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCurrentUser(t *testing.T) {
	e := echo.New()

	e.GET("/test", func(c echo.Context) error {
		claims, ok := httpEcho.GetCurrentUser(c)
		assert.False(t, ok)
		assert.Nil(t, claims)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestSkipPathMatching(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		skipPaths []string
		want      bool
	}{
		{
			name:      "exact match",
			path:      "/health",
			skipPaths: []string{"/health"},
			want:      true,
		},
		{
			name:      "prefix match with trailing slash",
			path:      "/api/users",
			skipPaths: []string{"/api/"},
			want:      true,
		},
		{
			name:      "prefix match without trailing slash",
			path:      "/api/users",
			skipPaths: []string{"/api"},
			want:      true,
		},
		{
			name:      "no false positive on prefix",
			path:      "/apifoo",
			skipPaths: []string{"/api"},
			want:      false,
		},
		{
			name:      "no match",
			path:      "/protected",
			skipPaths: []string{"/health"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import shouldSkipPath from echo package
			// For testing, we'll test through JWTMiddleware behavior
			config := &jwt.MiddlewareConfig{
				Config: jwt.DefaultConfig(),
			}
			config.Secret = "test-secret"
			config.TokenExtractor = jwt.DefaultTokenExtractor
			config.ErrorHandler = jwt.DefaultErrorHandler
			config.SkipPaths = tt.skipPaths

			e := echo.New()
			e.Use(httpEcho.JWTMiddleware(config))

			called := false
			e.GET(tt.path, func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// If should skip, handler should be called (200 or 204)
			// If should not skip, JWT should reject (401)
			if tt.want {
				assert.True(t, called, "Handler should have been called for skipped path")
			} else {
				assert.False(t, called, "Handler should not have been called without token")
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			}
		})
	}
}
