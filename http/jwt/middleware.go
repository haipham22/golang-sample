package jwt

import (
	"context"
	"net/http"
	"strings"

	"github.com/haipham22/govern/errors"
)

// ContextKey type for context keys.
type ContextKey string

// UserKey is the context key for user claims.
const UserKey ContextKey = "user"

// MiddlewareConfig configures JWT middleware.
type MiddlewareConfig struct {
	*Config
	TokenExtractor func(*http.Request) (string, error)
	ErrorHandler   func(http.ResponseWriter, *http.Request, error)
	SkipPaths      []string
}

// Middleware creates JWT authentication middleware.
func Middleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check skip paths
			if shouldSkipPath(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token
			tokenString, err := config.TokenExtractor(r)
			if err != nil {
				config.ErrorHandler(w, r, ErrTokenMissing)
				return
			}

			// Validate token
			claims, err := ValidateToken(tokenString, config.Config)
			if err != nil {
				config.ErrorHandler(w, r, err)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DefaultTokenExtractor extracts token from Authorization header.
func DefaultTokenExtractor(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrTokenMissing
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrTokenInvalid
	}

	if parts[1] == "" {
		return "", ErrTokenMissing
	}

	return parts[1], nil
}

// DefaultErrorHandler writes error responses.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if errors.IsCode(err, errors.CodeUnauthorized) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, err.Error(), http.StatusUnauthorized)
	} else {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
	}
}

// GetCurrentUser retrieves user from request context.
func GetCurrentUser(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(UserKey).(*Claims)
	return claims, ok
}

// MustGetCurrentUser retrieves user or panics.
func MustGetCurrentUser(r *http.Request) *Claims {
	claims, ok := GetCurrentUser(r)
	if !ok {
		panic("user not found in context")
	}
	return claims
}

// shouldSkipPath checks if path should be skipped.
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}
