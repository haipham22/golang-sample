package middlewares

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	// AllowOrigins is a list of allowed origins
	AllowOrigins []string
	// AllowMethods is a list of allowed methods
	AllowMethods []string
	// AllowHeaders is a list of allowed headers
	AllowHeaders []string
	// ExposeHeaders is a list of exposed headers
	ExposeHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge is the maximum age for preflight requests
	MaxAge int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.PATCH,
			echo.DELETE,
			echo.OPTIONS,
			echo.HEAD,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns CORS configuration for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORS returns a CORS middleware with default development config
func CORS() echo.MiddlewareFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) echo.MiddlewareFunc {
	corsConfig := echomiddleware.CORSConfig{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}

	return echomiddleware.CORSWithConfig(corsConfig)
}
