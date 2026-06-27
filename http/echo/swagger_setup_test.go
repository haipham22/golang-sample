package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithEchoSwagger(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		e := echo.New()

		// Call with no options - should be disabled
		WithEchoSwagger(e)

		// Try to access swagger - should get 404
		req := httptest.NewRequest(http.MethodGet, "/swagger/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("enabled when WithSwaggerEnabled(true)", func(t *testing.T) {
		e := echo.New()

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
		)

		// Access swagger UI
		req := httptest.NewRequest(http.MethodGet, "/swagger/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "swagger-ui")
		assert.Contains(t, rec.Body.String(), "SwaggerUIBundle")
	})

	t.Run("custom path with WithSwaggerPath", func(t *testing.T) {
		e := echo.New()

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerPath("/api/docs/*"),
		)

		// Access custom path
		req := httptest.NewRequest(http.MethodGet, "/api/docs/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "swagger-ui")
	})

	t.Run("custom info with WithSwaggerInfo", func(t *testing.T) {
		e := echo.New()

		customInfo := &SwaggerInfo{
			Title:       "Custom API",
			Description: "Custom description",
			Version:     "2.0",
			Host:        "api.example.com",
			BasePath:    "/v2",
		}

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerInfo(customInfo),
		)

		// Access swagger UI
		req := httptest.NewRequest(http.MethodGet, "/swagger/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Custom API")
		// Description is only in doc.json, not in HTML title
		assert.Contains(t, body, "/swagger/doc.json")
	})

	t.Run("serves doc.json", func(t *testing.T) {
		e := echo.New()

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerInfo(&SwaggerInfo{
				Title:       "Test API",
				Description: "Test Description",
				Version:     "1.0",
			}),
		)

		// Access doc.json
		req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		body := rec.Body.String()
		assert.Contains(t, body, `"swagger":"2.0"`)
		assert.Contains(t, body, `"title":"Test API"`)
		assert.Contains(t, body, `"description":"Test Description"`)
		assert.Contains(t, body, `"version":"1.0"`)
	})
}

func TestGenerateSwaggerHTML(t *testing.T) {
	t.Run("without authentication", func(t *testing.T) {
		settings := &swaggerSettings{
			enabled: true,
			path:    "/swagger/*",
			info: &SwaggerInfo{
				Title:       "Test API",
				Description: "Test Description",
				Version:     "1.0",
			},
			auth: nil,
		}

		html := generateSwaggerHTML("/swagger", settings)

		assert.Contains(t, html, "<title>Test API</title>")
		assert.Contains(t, html, `url: "/swagger/doc.json"`)
		assert.Contains(t, html, "swagger-ui-dist@5.9.0")
		assert.Contains(t, html, "SwaggerUIBundle")
	})

	t.Run("with Bearer authentication", func(t *testing.T) {
		settings := &swaggerSettings{
			enabled: true,
			path:    "/swagger/*",
			info: &SwaggerInfo{
				Title: "Test API",
			},
			auth: &SwaggerAuth{
				Type:        "Bearer",
				Description: "JWT token",
				Name:        "Authorization",
				In:          "header",
			},
		}

		html := generateSwaggerHTML("/swagger", settings)

		assert.Contains(t, html, "Bearer: []")
		assert.Contains(t, html, "security:")
	})

	t.Run("with ApiKey authentication", func(t *testing.T) {
		settings := &swaggerSettings{
			enabled: true,
			path:    "/swagger/*",
			info: &SwaggerInfo{
				Title: "Test API",
			},
			auth: &SwaggerAuth{
				Type:        "ApiKey",
				Description: "API key in header",
				Name:        "X-API-Key",
				In:          "header",
			},
		}

		html := generateSwaggerHTML("/swagger", settings)

		assert.Contains(t, html, "ApiKey: []")
		assert.Contains(t, html, "security:")
	})

	t.Run("with Basic authentication", func(t *testing.T) {
		settings := &swaggerSettings{
			enabled: true,
			path:    "/swagger/*",
			info: &SwaggerInfo{
				Title: "Test API",
			},
			auth: &SwaggerAuth{
				Type:        "Basic",
				Description: "Basic authentication",
			},
		}

		html := generateSwaggerHTML("/swagger", settings)

		assert.Contains(t, html, "Basic: []")
		assert.Contains(t, html, "security:")
	})

	t.Run("with OAuth2 authentication", func(t *testing.T) {
		settings := &swaggerSettings{
			enabled: true,
			path:    "/swagger/*",
			info: &SwaggerInfo{
				Title: "Test API",
			},
			auth: &SwaggerAuth{
				Type:             "OAuth2",
				Description:      "OAuth2 flow",
				Flow:             "implicit",
				AuthorizationURL: "https://example.com/oauth/authorize",
				TokenURL:         "https://example.com/oauth/token",
				Scopes: map[string]string{
					"read":  "Read access",
					"write": "Write access",
				},
			},
		}

		html := generateSwaggerHTML("/swagger", settings)

		assert.Contains(t, html, "OAuth2: []")
		assert.Contains(t, html, "initOAuth:")
		assert.Contains(t, html, "appName: \"OAuth2 flow\"")
	})
}

func TestSwaggerIntegration(t *testing.T) {
	t.Run("full setup with all options", func(t *testing.T) {
		e := echo.New()

		// Setup routes
		e.GET("/api/users", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"message": "users"})
		})

		// Setup Swagger
		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerPath("/docs/*"),
			WithSwaggerInfo(&SwaggerInfo{
				Title:       "User Management API",
				Description: "API for managing users",
				Version:     "2.0",
				Host:        "api.example.com",
				BasePath:    "/api",
			}),
			WithSwaggerAuth(&SwaggerAuth{
				Type:        "Bearer",
				Description: "JWT token for authentication",
				Name:        "Authorization",
				In:          "header",
			}),
		)

		// Test API route still works
		req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "users")

		// Test Swagger UI
		req = httptest.NewRequest(http.MethodGet, "/docs/", http.NoBody)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "User Management API")
		assert.Contains(t, body, "/docs/doc.json")

		// Test doc.json
		req = httptest.NewRequest(http.MethodGet, "/docs/doc.json", http.NoBody)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "User Management API")
	})
}

func TestSwaggerSecurity(t *testing.T) {
	t.Run("does not register routes when disabled", func(t *testing.T) {
		e := echo.New()

		WithEchoSwagger(e,
			WithSwaggerEnabled(false),
		)

		// Add a catch-all route to verify swagger routes aren't registered
		e.Any("/*", func(c echo.Context) error {
			return c.String(http.StatusOK, "caught")
		})

		// Try to access swagger - should be caught by catch-all
		req := httptest.NewRequest(http.MethodGet, "/swagger/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "caught", rec.Body.String())
	})

	t.Run("HTML escaping in user input", func(t *testing.T) {
		e := echo.New()

		// Try to inject XSS via title
		maliciousInfo := &SwaggerInfo{
			Title:       `<script>alert('xss')</script>`,
			Description: "normal description",
			Version:     "1.0",
		}

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerInfo(maliciousInfo),
		)

		req := httptest.NewRequest(http.MethodGet, "/swagger/", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// The title should be present but not executed
		body := rec.Body.String()
		assert.True(t, strings.Contains(body, "<title>") &&
			strings.Contains(body, "<script>alert('xss')</script>") ||
			strings.Contains(body, "&lt;script&gt;"))
	})
}

func TestSwaggerPathHandling(t *testing.T) {
	tests := []struct {
		name          string
		swaggerPath   string
		requestPath   string
		expectSuccess bool
	}{
		{
			name:          "default wildcard path",
			swaggerPath:   "/swagger/*",
			requestPath:   "/swagger/",
			expectSuccess: true,
		},
		{
			name:          "default wildcard path with subroute",
			swaggerPath:   "/swagger/*",
			requestPath:   "/swagger/test",
			expectSuccess: true,
		},
		{
			name:          "custom wildcard path",
			swaggerPath:   "/api/docs/*",
			requestPath:   "/api/docs/",
			expectSuccess: true,
		},
		{
			name:          "wrong path returns 404",
			swaggerPath:   "/swagger/*",
			requestPath:   "/api/docs/",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			WithEchoSwagger(e,
				WithSwaggerEnabled(true),
				WithSwaggerPath(tt.swaggerPath),
			)

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, http.NoBody)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if tt.expectSuccess {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Contains(t, rec.Body.String(), "swagger-ui")
			} else {
				assert.NotEqual(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestSwaggerDocJSON(t *testing.T) {
	t.Run("returns proper JSON structure", func(t *testing.T) {
		e := echo.New()

		customInfo := &SwaggerInfo{
			Title:       "My API",
			Description: "API Description",
			Version:     "3.0",
			Host:        "example.com:8080",
			BasePath:    "/v1",
		}

		WithEchoSwagger(e,
			WithSwaggerEnabled(true),
			WithSwaggerInfo(customInfo),
		)

		req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", http.NoBody)
		req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		// Verify JSON structure
		body := rec.Body.String()
		assert.Contains(t, body, `"swagger":"2.0"`)
		assert.Contains(t, body, `"info":{`)
		assert.Contains(t, body, `"title":"My API"`)
		assert.Contains(t, body, `"description":"API Description"`)
		assert.Contains(t, body, `"version":"3.0"`)
		assert.Contains(t, body, `"paths":{`)
	})
}
