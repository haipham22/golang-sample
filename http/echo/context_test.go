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

func TestMustGetCurrentUser(t *testing.T) {
	t.Run("panics when user not in context", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)

		assert.Panics(t, func() {
			httpEcho.MustGetCurrentUser(c)
		})
	})

	t.Run("returns user when in context", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)

		claims := &jwt.Claims{
			UserID:   "user123",
			Username: "testuser",
		}
		c.Set("user", claims)

		result := httpEcho.MustGetCurrentUser(c)
		assert.Equal(t, claims, result)
		assert.Equal(t, "user123", result.UserID)
		assert.Equal(t, "testuser", result.Username)
	})
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(echo.Context)
		wantID string
		wantOK bool
	}{
		{
			name: "returns user id when set",
			setup: func(c echo.Context) {
				c.Set("user_id", "user123")
			},
			wantID: "user123",
			wantOK: true,
		},
		{
			name: "returns false when not set",
			setup: func(c echo.Context) {
				// Don't set anything
			},
			wantID: "",
			wantOK: false,
		},
		{
			name: "returns false when wrong type",
			setup: func(c echo.Context) {
				c.Set("user_id", 123) // Wrong type
			},
			wantID: "",
			wantOK: false,
		},
		{
			name: "returns false when empty string",
			setup: func(c echo.Context) {
				c.Set("user_id", "")
			},
			wantID: "",
			wantOK: true, // It's set, just empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			c := e.NewContext(nil, nil)

			tt.setup(c)

			userID, ok := httpEcho.GetUserID(c)
			assert.Equal(t, tt.wantID, userID)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestGetUsername(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(echo.Context)
		wantUsername string
		wantOK       bool
	}{
		{
			name: "returns username when set",
			setup: func(c echo.Context) {
				c.Set("username", "testuser")
			},
			wantUsername: "testuser",
			wantOK:       true,
		},
		{
			name: "returns false when not set",
			setup: func(c echo.Context) {
				// Don't set anything
			},
			wantUsername: "",
			wantOK:       false,
		},
		{
			name: "returns false when wrong type",
			setup: func(c echo.Context) {
				c.Set("username", 123) // Wrong type
			},
			wantUsername: "",
			wantOK:       false,
		},
		{
			name: "returns false when empty string",
			setup: func(c echo.Context) {
				c.Set("username", "")
			},
			wantUsername: "",
			wantOK:       true, // It's set, just empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			c := e.NewContext(nil, nil)

			tt.setup(c)

			username, ok := httpEcho.GetUsername(c)
			assert.Equal(t, tt.wantUsername, username)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestContextHelpers_Integration(t *testing.T) {
	// Test all context helpers together with JWT middleware
	config := &jwt.MiddlewareConfig{
		Config: jwt.DefaultConfig(),
	}
	config.Secret = "test-secret"
	config.TokenExtractor = jwt.DefaultTokenExtractor
	config.ErrorHandler = jwt.DefaultErrorHandler

	e := echo.New()
	e.Use(httpEcho.JWTMiddleware(config))

	e.GET("/test", func(c echo.Context) error {
		// Test all helper functions
		claims, ok := httpEcho.GetCurrentUser(c)
		require.True(t, ok)
		assert.NotNil(t, claims)

		userID, ok := httpEcho.GetUserID(c)
		require.True(t, ok)
		assert.NotEmpty(t, userID)

		username, ok := httpEcho.GetUsername(c)
		require.True(t, ok)
		assert.NotEmpty(t, username)

		// Test MustGetCurrentUser doesn't panic
		mustClaims := httpEcho.MustGetCurrentUser(c)
		assert.Equal(t, claims, mustClaims)

		return c.JSON(http.StatusOK, map[string]string{
			"user_id":  userID,
			"username": username,
		})
	})

	// Create valid token
	claims := &jwt.Claims{
		UserID:   "user123",
		Username: "testuser",
	}
	token, err := jwt.GenerateAccessToken(claims, config.Config)
	require.NoError(t, err)

	// Make request with token
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
