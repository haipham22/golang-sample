package echo

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/haipham22/govern/http/jwt"
)

// JWTMiddleware creates JWT authentication middleware for Echo.
func JWTMiddleware(config *jwt.MiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check skip paths with proper prefix matching
			if shouldSkipPath(c.Path(), config.SkipPaths) {
				return next(c)
			}

			// Extract token using Echo request
			tokenString, err := config.TokenExtractor(c.Request())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			// Validate token
			claims, err := jwt.ValidateToken(tokenString, config.Config)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			// Store in Echo context
			c.Set("user", claims)
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			return next(c)
		}
	}
}

// GetCurrentUser retrieves user from Echo context.
// Returns claims and true if found, nil and false otherwise.
func GetCurrentUser(c echo.Context) (*jwt.Claims, bool) {
	claims, ok := c.Get("user").(*jwt.Claims)
	return claims, ok
}

// MustGetCurrentUser retrieves user or panics.
// Use this in handlers that are guaranteed to have JWT middleware.
func MustGetCurrentUser(c echo.Context) *jwt.Claims {
	claims, ok := GetCurrentUser(c)
	if !ok {
		panic("user not found in context")
	}
	return claims
}

// GetUserID retrieves user ID from context.
func GetUserID(c echo.Context) (string, bool) {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID, true
	}
	return "", false
}

// GetUsername retrieves username from context.
func GetUsername(c echo.Context) (string, bool) {
	if username, ok := c.Get("username").(string); ok {
		return username, true
	}
	return "", false
}

// shouldSkipPath checks if path should skip JWT validation.
// Uses proper prefix matching to avoid false positives.
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		// Exact match
		if path == skip {
			return true
		}

		// Prefix match with proper boundary checking
		// Only match if:
		// - skip ends with / and path starts with skip (e.g., /api/ matches /api/foo)
		// - skip doesn't end with / and path has / after skip (e.g., /api matches /api/foo but not /apifoo)
		if strings.HasSuffix(skip, "/") {
			if strings.HasPrefix(path, skip) {
				return true
			}
		} else {
			if strings.HasPrefix(path, skip+"/") {
				return true
			}
		}
	}
	return false
}
