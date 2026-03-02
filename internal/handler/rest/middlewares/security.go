package middlewares

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders returns a middleware that adds security headers to all responses
func SecurityHeaders() echo.MiddlewareFunc {
	return SecurityHeadersWithConfig(DefaultSecurityHeadersConfig())
}

// SecurityHeadersConfig holds configuration for security headers middleware
type SecurityHeadersConfig struct {
	// FrameOptions controls X-Frame-Options header
	FrameOptions string
	// HSTSMaxAge is the max-age for Strict-Transport-Security
	HSTSMaxAge int
	// CSP is the Content-Security-Policy header value
	CSP string
	// HSTSIncludeSubDomains indicates whether to include subdomains in HSTS
	HSTSIncludeSubDomains bool
}

// DefaultSecurityHeadersConfig returns default security headers configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		FrameOptions:          "DENY",
		HSTSMaxAge:            31536000,
		CSP:                   "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';",
		HSTSIncludeSubDomains: true,
	}
}

// SecurityHeadersWithConfig returns a security headers middleware with custom configuration
func SecurityHeadersWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// X-Frame-Options
			if config.FrameOptions != "" {
				c.Response().Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-Content-Type-Options
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict-Transport-Security
			// Use safe default if not configured (63072000 seconds = 2 years)
			maxAge := config.HSTSMaxAge
			if maxAge == 0 {
				maxAge = 63072000
			}
			hstsValue := fmt.Sprintf("max-age=%d", maxAge)
			if config.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}
			c.Response().Header().Set("Strict-Transport-Security", hstsValue)

			// Content-Security-Policy
			if config.CSP != "" {
				c.Response().Header().Set("Content-Security-Policy", config.CSP)
			}

			// Referrer-Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy
			c.Response().Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			return next(c)
		}
	}
}
