package echo

// ExampleWithSwaggerAuth_bearerToken shows how to configure Swagger UI
// with Bearer token authentication (JWT).
//
// In your main.go, add these annotations for swag:
//
//	// @securityDefinitions.apikey Bearer
//	// @in header
//	// @name Authorization
//	// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"
func ExampleWithSwaggerAuth_bearerToken() {
	// Create swagger config with Bearer token authentication
	swaggerConfig := &swaggerSettings{
		enabled: true,
		auth: &SwaggerAuth{
			Type:        "Bearer",
			Description: "JWT token",
			Name:        "Authorization",
			In:          "header",
		},
	}

	dummyOperation(swaggerConfig)
}

// ExampleWithSwaggerAuth_apiKey shows API key authentication.
func ExampleWithSwaggerAuth_apiKey() {
	// Create swagger config with API key authentication
	swaggerConfig := &swaggerSettings{
		enabled: true,
		auth: &SwaggerAuth{
			Type:        "ApiKey",
			Description: "API key authentication",
			Name:        "X-API-Key",
			In:          "header",
		},
	}

	dummyOperation(swaggerConfig)
}

// ExampleWithSwaggerAuth_basicAuth shows Basic authentication.
func ExampleWithSwaggerAuth_basicAuth() {
	// Create swagger config with Basic authentication
	swaggerConfig := &swaggerSettings{
		enabled: true,
		auth: &SwaggerAuth{
			Type:        "Basic",
			Description: "Basic authentication",
		},
	}

	dummyOperation(swaggerConfig)
}

// ExampleWithSwaggerAuth_oauth2 shows OAuth2 authentication.
func ExampleWithSwaggerAuth_oauth2() {
	// Create swagger config with OAuth2 authentication
	swaggerConfig := &swaggerSettings{
		enabled: true,
		auth: &SwaggerAuth{
			Type:             "OAuth2",
			Flow:             "accessCode",
			AuthorizationURL: "https://example.com/oauth/authorize",
			TokenURL:         "https://example.com/oauth/token",
			Scopes: map[string]string{
				"read":  "Read access",
				"write": "Write access",
			},
		},
	}

	dummyOperation(swaggerConfig)
}

func dummyOperation(config *swaggerSettings) {
	_ = config
}
