package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds JWT configuration.
type Config struct {
	Secret          string
	SigningMethod   jwt.SigningMethod
	AccessDuration  time.Duration
	RefreshDuration time.Duration
	Issuer          string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		SigningMethod:   jwt.SigningMethodHS256,
		AccessDuration:  15 * time.Minute,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "govern",
	}
}

// Validate checks configuration validity.
func (c *Config) Validate() error {
	if c.Secret == "" {
		return ErrSecretRequired
	}
	if c.SigningMethod == nil {
		return ErrSigningMethodRequired
	}
	if c.AccessDuration <= 0 {
		return ErrInvalidAccessDuration
	}
	if c.RefreshDuration <= 0 {
		return ErrInvalidRefreshDuration
	}
	return nil
}

// GenerateSecureSecret generates a cryptographically secure secret.
func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAccessToken creates a new access token.
func GenerateAccessToken(claims *Claims, config *Config) (string, error) {
	if err := config.Validate(); err != nil {
		return "", err
	}

	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    config.Issuer,
		Subject:   claims.UserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessDuration)),
	}

	token := jwt.NewWithClaims(config.SigningMethod, claims)
	return token.SignedString([]byte(config.Secret))
}

// GenerateRefreshToken creates a new refresh token.
func GenerateRefreshToken(claims *Claims, config *Config) (string, error) {
	if err := config.Validate(); err != nil {
		return "", err
	}

	refreshClaims := &Claims{
		UserID: claims.UserID,
	}

	now := time.Now()
	refreshClaims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    config.Issuer,
		Subject:   claims.UserID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshDuration)),
	}

	token := jwt.NewWithClaims(config.SigningMethod, refreshClaims)
	return token.SignedString([]byte(config.Secret))
}

// ValidateToken validates a JWT token and returns claims.
func ValidateToken(tokenString string, config *Config) (*Claims, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method != config.SigningMethod {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return nil, HandleJWTError(err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	if err := claims.Validate(); err != nil {
		return nil, err
	}

	return claims, nil
}

// RefreshAccessToken generates new access token from refresh token.
func RefreshAccessToken(refreshToken string, config *Config) (string, error) {
	claims, err := ValidateToken(refreshToken, config)
	if err != nil {
		return "", err
	}

	return GenerateAccessToken(claims, config)
}
