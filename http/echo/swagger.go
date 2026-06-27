package echo

// SwaggerOption configures Swagger UI settings.
type SwaggerOption func(*swaggerSettings)

// swaggerSettings holds internal Swagger configuration.
type swaggerSettings struct {
	enabled bool
	path    string
	info    *SwaggerInfo
	auth    *SwaggerAuth
}

// WithSwaggerEnabled enables or disables Swagger UI.
// Default: false (disabled for security).
//
// Security Warning: Only enable in development environments or behind authentication.
func WithSwaggerEnabled(enabled bool) SwaggerOption {
	return func(s *swaggerSettings) {
		s.enabled = enabled
	}
}

// WithSwaggerPath sets custom route path for Swagger UI.
// Default: "/swagger/*"
//
// The path must end with "/*" to match all sub-routes.
func WithSwaggerPath(path string) SwaggerOption {
	return func(s *swaggerSettings) {
		s.path = path
	}
}

// WithSwaggerInfo sets API metadata for Swagger documentation.
func WithSwaggerInfo(info *SwaggerInfo) SwaggerOption {
	return func(s *swaggerSettings) {
		s.info = info
	}
}

// WithSwaggerAuth configures authentication for Swagger UI.
//
// This enables the "Authorize" button in Swagger UI, allowing users to test
// authenticated endpoints. The authentication configuration is displayed in
// the Swagger UI and can be used for try-it-out requests.
//
// Example:
//
//	server := echo.NewServer(":8080",
//	    echo.WithEchoSwagger(
//	        echo.WithSwaggerEnabled(true),
//	        echo.WithSwaggerAuth(&echo.SwaggerAuth{
//	            Type:        "Bearer",
//	            Description: "JWT token",
//	            Name:        "Authorization",
//	            In:          "header",
//	        }),
//	    ),
//	)
func WithSwaggerAuth(auth *SwaggerAuth) SwaggerOption {
	return func(s *swaggerSettings) {
		s.auth = auth
	}
}

// SwaggerInfo contains API metadata for Swagger documentation.
//
// These fields correspond to Swagger/OpenAPI specification fields.
// They are used to generate the documentation header.
type SwaggerInfo struct {
	// Title is the API title (default: "API Documentation").
	Title string

	// Description is a short description of the API.
	Description string

	// Version is the API version (default: "1.0").
	Version string

	// Host is the host name or IP (default: "localhost:8080").
	Host string

	// BasePath is the base path for all API routes (default: "/").
	BasePath string

	// TermsOfService is the URL to terms of service.
	TermsOfService string

	// Contact information
	ContactName  string
	ContactURL   string
	ContactEmail string

	// License information
	LicenseName string
	LicenseURL  string

	// Schemes are the supported protocols (http, https, ws, wss).
	// Default: []string{"http", "https"}
	Schemes []string
}

// DefaultSwaggerInfo provides default values for SwaggerInfo.
var DefaultSwaggerInfo = &SwaggerInfo{
	Title:       "API Documentation",
	Description: "This is a sample server.",
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/",
	Schemes:     []string{"http", "https"},
}

// SwaggerAuth configures authentication for Swagger UI.
//
// This defines how the "Authorize" button in Swagger UI will work,
// allowing users to test authenticated endpoints.
//
// Common configurations:
//
//	// Bearer token (JWT)
//	&SwaggerAuth{
//	    Type:        "Bearer",
//	    Description: "JWT token",
//	    Name:        "Authorization",
//	    In:          "header",
//	}
//
//	// API key
//	&SwaggerAuth{
//	    Type:        "ApiKey",
//	    Description: "API key",
//	    Name:        "X-API-Key",
//	    In:          "header",
//	}
//
//	// Basic auth
//	&SwaggerAuth{
//	    Type:        "Basic",
//	    Description: "Basic authentication",
//	}
type SwaggerAuth struct {
	// Type is the authentication type.
	// Common values: "Bearer", "ApiKey", "Basic", "OAuth2"
	Type string

	// Description describes the authentication scheme.
	Description string

	// Name is the header or query parameter name.
	// For "Bearer" with "header", use "Authorization".
	// For API keys, use your custom header name.
	Name string

	// In is where the credential is passed.
	// Values: "header" or "query".
	// Default: "header"
	In string

	// Flow is the OAuth2 flow type (only for OAuth2).
	// Values: "implicit", "password", "application", "accessCode".
	Flow string

	// AuthorizationURL is the OAuth2 authorization URL.
	AuthorizationURL string

	// TokenURL is the OAuth2 token URL.
	TokenURL string

	// Scopes are the available OAuth2 scopes.
	// Map of scope name to description.
	Scopes map[string]string
}
