package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT custom claims.
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// Validate performs claims validation.
func (c *Claims) Validate() error {
	if c.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// HasRole checks if user has role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles.
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}
