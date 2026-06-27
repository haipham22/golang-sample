// Package echo provides utilities for using Echo framework with Govern.
//
// This package provides JWT authentication middleware, Swagger UI integration,
// and handler wrapping utilities for Echo applications.
//
// # JWT Authentication
//
// Create JWT middleware for Echo routes:
//
//	jwtConfig := &jwt.MiddlewareConfig{
//	    Config:         jwt.DefaultConfig(),
//	    TokenExtractor: jwt.DefaultTokenExtractor,
//	    SkipPaths:      []string{"/health", "/login"},
//	}
//	jwtConfig.Config.Secret = "your-secret-key"
//
//	// Use with Echo
//	e := echo.New()
//	e.Use(echo.JWTMiddleware(jwtConfig))
//
// # Getting Current User
//
//	func handler(c echo.Context) error {
//	    claims, ok := echo.GetCurrentUser(c)
//	    if !ok {
//	        return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
//	    }
//
//	    // Or panic if not found (use after JWT middleware)
//	    claims := echo.MustGetCurrentUser(c)
//
//	    return c.JSON(http.StatusOK, claims)
//	}
//
// # Swagger UI Integration
//
// This package supports Swagger UI for interactive API documentation.
//
// ## Setup
//
// 1. Install swag CLI:
//
//	go install github.com/swaggo/swag/cmd/swag@latest
//
// 2. Add Swagger annotations to your handlers:
//
//	// @Summary Get user by ID
//	// @Description Retrieve user information
//	// @Tags users
//	// @Accept json
//	// @Produce json
//	// @Param id path int true "User ID"
//	// @Success 200 {object} User
//	// @Failure 404 {object} ErrorResponse
//	// @Router /users/{id} [get]
//	func getUser(c echo.Context) error {
//	    // Handler logic
//	}
//
// 3. Generate Swagger docs:
//
//	swag init -g cmd/api/main.go
//
// 4. Configure Swagger UI options:
//
//	import _ "myapi/docs"  // Generated docs
//
//	swaggerConfig := echo.WithSwaggerEnabled(true)
//	swaggerConfig = echo.WithSwaggerInfo(&echo.SwaggerInfo{
//	    Title:       "My API",
//	    Description: "Sample API server",
//	    Version:     "1.0",
//	})(swaggerConfig)
//
// 5. Apply to Echo instance manually
//
// ## Security
//
// Swagger UI should not be exposed in production without authentication.
// Use WithSwaggerEnabled with environment variables:
//
//	enableSwagger := os.Getenv("GO_ENV") == "development"
//	swaggerConfig := echo.WithSwaggerEnabled(enableSwagger)
//
// ## With JWT Authentication
//
// To enable Bearer token authentication in Swagger UI:
//
//	swaggerConfig := echo.WithSwaggerAuth(&echo.SwaggerAuth{
//	    Type:        "Bearer",
//	    Description: "JWT token",
//	    Name:        "Authorization",
//	    In:          "header",
//	})
//
// Add these annotations to main.go:
//
//	// @securityDefinitions.apikey Bearer
//	// @in header
//	// @name Authorization
//	// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"
//
// # Handler Wrapping
//
// Wrap standard http.Handler for use with Echo:
//
//	var myHandler http.Handler = // ...
//	e.GET("/legacy", echo.WrapHandler(myHandler))
//
// # Middleware Compatibility
//
// IMPORTANT: Echo middleware (echo.MiddlewareFunc) and Govern middleware (http.Middleware)
// are NOT compatible due to different error handling and context types.
// Use Echo middleware for Echo routes, and Govern middleware for http.Handler routes.
package echo
