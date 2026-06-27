package echo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	httpEcho "github.com/haipham22/govern/http/echo"
)

func TestSwaggerInfo(t *testing.T) {
	t.Run("DefaultSwaggerInfo has expected values", func(t *testing.T) {
		info := httpEcho.DefaultSwaggerInfo
		assert.NotNil(t, info)
		assert.Equal(t, "API Documentation", info.Title)
		assert.Equal(t, "This is a sample server.", info.Description)
		assert.Equal(t, "1.0", info.Version)
		assert.Equal(t, "localhost:8080", info.Host)
		assert.Equal(t, "/", info.BasePath)
		assert.Equal(t, []string{"http", "https"}, info.Schemes)
	})

	t.Run("custom SwaggerInfo can be created", func(t *testing.T) {
		info := &httpEcho.SwaggerInfo{
			Title:       "My API",
			Description: "My custom API",
			Version:     "2.0",
			Host:        "api.example.com",
			BasePath:    "/v1",
			Schemes:     []string{"https"},
		}

		assert.Equal(t, "My API", info.Title)
		assert.Equal(t, "My custom API", info.Description)
		assert.Equal(t, "2.0", info.Version)
		assert.Equal(t, "api.example.com", info.Host)
		assert.Equal(t, "/v1", info.BasePath)
		assert.Equal(t, []string{"https"}, info.Schemes)
	})

	t.Run("SwaggerInfo supports all optional fields", func(t *testing.T) {
		info := &httpEcho.SwaggerInfo{
			Title:          "Complete API",
			Description:    "Complete API description",
			Version:        "3.0",
			Host:           "example.com",
			BasePath:       "/api/v3",
			TermsOfService: "https://example.com/terms",
			ContactName:    "API Team",
			ContactURL:     "https://example.com/contact",
			ContactEmail:   "api@example.com",
			LicenseName:    "MIT",
			LicenseURL:     "https://opensource.org/licenses/MIT",
			Schemes:        []string{"http", "https", "ws", "wss"},
		}

		assert.Equal(t, "Complete API", info.Title)
		assert.Equal(t, "Complete API description", info.Description)
		assert.Equal(t, "3.0", info.Version)
		assert.Equal(t, "example.com", info.Host)
		assert.Equal(t, "/api/v3", info.BasePath)
		assert.Equal(t, "https://example.com/terms", info.TermsOfService)
		assert.Equal(t, "API Team", info.ContactName)
		assert.Equal(t, "https://example.com/contact", info.ContactURL)
		assert.Equal(t, "api@example.com", info.ContactEmail)
		assert.Equal(t, "MIT", info.LicenseName)
		assert.Equal(t, "https://opensource.org/licenses/MIT", info.LicenseURL)
		assert.Equal(t, []string{"http", "https", "ws", "wss"}, info.Schemes)
	})
}

func TestSwaggerAuth(t *testing.T) {
	t.Run("Bearer token configuration", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:        "Bearer",
			Description: "JWT token",
			Name:        "Authorization",
			In:          "header",
		}

		assert.Equal(t, "Bearer", auth.Type)
		assert.Equal(t, "JWT token", auth.Description)
		assert.Equal(t, "Authorization", auth.Name)
		assert.Equal(t, "header", auth.In)
	})

	t.Run("API key configuration", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:        "ApiKey",
			Description: "API key authentication",
			Name:        "X-API-Key",
			In:          "header",
		}

		assert.Equal(t, "ApiKey", auth.Type)
		assert.Equal(t, "API key authentication", auth.Description)
		assert.Equal(t, "X-API-Key", auth.Name)
		assert.Equal(t, "header", auth.In)
	})

	t.Run("API key in query parameter", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:        "ApiKey",
			Description: "API key in query",
			Name:        "api_key",
			In:          "query",
		}

		assert.Equal(t, "ApiKey", auth.Type)
		assert.Equal(t, "API key in query", auth.Description)
		assert.Equal(t, "query", auth.In)
		assert.Equal(t, "api_key", auth.Name)
	})

	t.Run("Basic auth configuration", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:        "Basic",
			Description: "Basic authentication",
		}

		assert.Equal(t, "Basic", auth.Type)
		assert.Equal(t, "Basic authentication", auth.Description)
		assert.Empty(t, auth.Name)
		assert.Empty(t, auth.In)
	})

	t.Run("OAuth2 configuration with all fields", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:             "OAuth2",
			Flow:             "accessCode",
			AuthorizationURL: "https://example.com/oauth/authorize",
			TokenURL:         "https://example.com/oauth/token",
			Scopes: map[string]string{
				"read":  "Read access",
				"write": "Write access",
				"admin": "Admin access",
			},
		}

		assert.Equal(t, "OAuth2", auth.Type)
		assert.Equal(t, "accessCode", auth.Flow)
		assert.Equal(t, "https://example.com/oauth/authorize", auth.AuthorizationURL)
		assert.Equal(t, "https://example.com/oauth/token", auth.TokenURL)
		assert.Len(t, auth.Scopes, 3)
		assert.Equal(t, "Read access", auth.Scopes["read"])
		assert.Equal(t, "Write access", auth.Scopes["write"])
		assert.Equal(t, "Admin access", auth.Scopes["admin"])
	})

	t.Run("OAuth2 with implicit flow", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:             "OAuth2",
			Flow:             "implicit",
			AuthorizationURL: "https://example.com/oauth/authorize",
			Scopes: map[string]string{
				"read": "Read access",
			},
		}

		assert.Equal(t, "OAuth2", auth.Type)
		assert.Equal(t, "implicit", auth.Flow)
		assert.Equal(t, "https://example.com/oauth/authorize", auth.AuthorizationURL)
		assert.Len(t, auth.Scopes, 1)
		assert.Equal(t, "Read access", auth.Scopes["read"])
		assert.Empty(t, auth.TokenURL) // Not needed for implicit flow
	})

	t.Run("OAuth2 with password flow", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:     "OAuth2",
			Flow:     "password",
			TokenURL: "https://example.com/oauth/token",
			Scopes: map[string]string{
				"full": "Full access",
			},
		}

		assert.Equal(t, "OAuth2", auth.Type)
		assert.Equal(t, "password", auth.Flow)
		assert.Equal(t, "https://example.com/oauth/token", auth.TokenURL)
		assert.Len(t, auth.Scopes, 1)
		assert.Equal(t, "Full access", auth.Scopes["full"])
	})

	t.Run("OAuth2 with application flow", func(t *testing.T) {
		auth := &httpEcho.SwaggerAuth{
			Type:     "OAuth2",
			Flow:     "application",
			TokenURL: "https://example.com/oauth/token",
			Scopes: map[string]string{
				"service": "Service account access",
			},
		}

		assert.Equal(t, "OAuth2", auth.Type)
		assert.Equal(t, "application", auth.Flow)
		assert.Equal(t, "https://example.com/oauth/token", auth.TokenURL)
		assert.Len(t, auth.Scopes, 1)
		assert.Equal(t, "Service account access", auth.Scopes["service"])
	})
}

func TestSwaggerOptions(t *testing.T) {
	t.Run("WithSwaggerEnabled enables swagger", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		option := httpEcho.WithSwaggerEnabled(true)
		settings.ExportApplySwaggerOption(option)

		assert.True(t, settings.ExportGetSwaggerEnabled())
	})

	t.Run("WithSwaggerEnabled disables swagger", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		option := httpEcho.WithSwaggerEnabled(false)
		settings.ExportApplySwaggerOption(option)

		assert.False(t, settings.ExportGetSwaggerEnabled())
	})

	t.Run("WithSwaggerPath sets custom path", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		option := httpEcho.WithSwaggerPath("/docs/*")
		settings.ExportApplySwaggerOption(option)

		assert.Equal(t, "/docs/*", settings.ExportGetSwaggerPath())
	})

	t.Run("WithSwaggerPath with default", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		option := httpEcho.WithSwaggerPath("/swagger/*")
		settings.ExportApplySwaggerOption(option)

		assert.Equal(t, "/swagger/*", settings.ExportGetSwaggerPath())
	})

	t.Run("WithSwaggerInfo sets info", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		info := &httpEcho.SwaggerInfo{
			Title:       "Test API",
			Description: "Test description",
			Version:     "1.0",
		}
		option := httpEcho.WithSwaggerInfo(info)
		settings.ExportApplySwaggerOption(option)

		retrievedInfo := settings.ExportGetSwaggerInfo()
		assert.Equal(t, info, retrievedInfo)
		assert.Equal(t, "Test API", retrievedInfo.Title)
		assert.Equal(t, "Test description", retrievedInfo.Description)
		assert.Equal(t, "1.0", retrievedInfo.Version)
	})

	t.Run("WithSwaggerInfo with DefaultSwaggerInfo", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		option := httpEcho.WithSwaggerInfo(httpEcho.DefaultSwaggerInfo)
		settings.ExportApplySwaggerOption(option)

		retrievedInfo := settings.ExportGetSwaggerInfo()
		assert.Equal(t, httpEcho.DefaultSwaggerInfo, retrievedInfo)
		assert.Equal(t, "API Documentation", retrievedInfo.Title)
	})

	t.Run("WithSwaggerAuth sets auth", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()
		auth := &httpEcho.SwaggerAuth{
			Type:        "Bearer",
			Description: "JWT",
			Name:        "Authorization",
			In:          "header",
		}
		option := httpEcho.WithSwaggerAuth(auth)
		settings.ExportApplySwaggerOption(option)

		retrievedAuth := settings.ExportGetSwaggerAuth()
		assert.Equal(t, auth, retrievedAuth)
		assert.Equal(t, "Bearer", retrievedAuth.Type)
		assert.Equal(t, "JWT", retrievedAuth.Description)
		assert.Equal(t, "Authorization", retrievedAuth.Name)
		assert.Equal(t, "header", retrievedAuth.In)
	})

	t.Run("multiple options can be applied", func(t *testing.T) {
		settings := httpEcho.ExportGetSwaggerSettings()

		info := &httpEcho.SwaggerInfo{
			Title:       "Multi Option API",
			Description: "Multi option test API",
			Version:     "2.0",
		}

		auth := &httpEcho.SwaggerAuth{
			Type:        "Bearer",
			Description: "Bearer auth",
		}

		settings.ExportApplySwaggerOption(httpEcho.WithSwaggerEnabled(true))
		settings.ExportApplySwaggerOption(httpEcho.WithSwaggerPath("/api/docs/*"))
		settings.ExportApplySwaggerOption(httpEcho.WithSwaggerInfo(info))
		settings.ExportApplySwaggerOption(httpEcho.WithSwaggerAuth(auth))

		assert.True(t, settings.ExportGetSwaggerEnabled())
		assert.Equal(t, "/api/docs/*", settings.ExportGetSwaggerPath())
		assert.Equal(t, info, settings.ExportGetSwaggerInfo())
		assert.Equal(t, auth, settings.ExportGetSwaggerAuth())
	})
}
