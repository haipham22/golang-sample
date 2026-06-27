package jwt

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goverrors "github.com/haipham22/govern/errors"
)

// Test helpers

func createTestConfig(t *testing.T) *Config {
	t.Helper()
	return &Config{
		Secret:          "test-secret-key-for-testing-purposes-only",
		SigningMethod:   jwt.SigningMethodHS256,
		AccessDuration:  15 * time.Minute,
		RefreshDuration: 7 * 24 * time.Hour,
		Issuer:          "govern-test",
	}
}

func createTestClaims(t *testing.T) *Claims {
	t.Helper()
	return &Claims{
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin", "user"},
	}
}

func assertTokenValid(t *testing.T, tokenString string, config *Config) {
	t.Helper()
	claims, err := ValidateToken(tokenString, config)
	require.NoError(t, err)
	require.NotNil(t, claims)
}

func assertErrorCode(t *testing.T, err, wantErr error) {
	t.Helper()
	assert.Error(t, err)
	assert.Equal(t, wantErr, err)
}

// Config tests

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantErr     error
		description string
	}{
		{
			name:        "valid config",
			config:      createTestConfig(t),
			wantErr:     nil,
			description: "should pass validation with all fields set",
		},
		{
			name: "missing secret",
			config: &Config{
				SigningMethod:   jwt.SigningMethodHS256,
				AccessDuration:  15 * time.Minute,
				RefreshDuration: 7 * 24 * time.Hour,
			},
			wantErr:     ErrSecretRequired,
			description: "should fail with empty secret",
		},
		{
			name: "missing signing method",
			config: &Config{
				Secret:          "test-secret",
				AccessDuration:  15 * time.Minute,
				RefreshDuration: 7 * 24 * time.Hour,
			},
			wantErr:     ErrSigningMethodRequired,
			description: "should fail with nil signing method",
		},
		{
			name: "invalid access duration",
			config: &Config{
				Secret:          "test-secret",
				SigningMethod:   jwt.SigningMethodHS256,
				AccessDuration:  0,
				RefreshDuration: 7 * 24 * time.Hour,
			},
			wantErr:     ErrInvalidAccessDuration,
			description: "should fail with zero access duration",
		},
		{
			name: "invalid refresh duration",
			config: &Config{
				Secret:          "test-secret",
				SigningMethod:   jwt.SigningMethodHS256,
				AccessDuration:  15 * time.Minute,
				RefreshDuration: -1 * time.Hour,
			},
			wantErr:     ErrInvalidRefreshDuration,
			description: "should fail with negative refresh duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, jwt.SigningMethodHS256, config.SigningMethod)
	assert.Equal(t, 15*time.Minute, config.AccessDuration)
	assert.Equal(t, 7*24*time.Hour, config.RefreshDuration)
	assert.Equal(t, "govern", config.Issuer)
}

// Token generation tests

func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		setupClaims func(*testing.T) *Claims
		setupConfig func(*testing.T) *Config
		validate    func(*testing.T, string, *Config)
		wantErr     error
		description string
	}{
		{
			name:        "valid token generation",
			setupClaims: createTestClaims,
			setupConfig: createTestConfig,
			validate: func(t *testing.T, token string, config *Config) {
				t.Helper()
				assertTokenValid(t, token, config)
			},
			wantErr:     nil,
			description: "should generate valid access token",
		},
		{
			name: "token with minimal claims",
			setupClaims: func(t *testing.T) *Claims {
				t.Helper()
				return &Claims{
					UserID: "user456",
				}
			},
			setupConfig: createTestConfig,
			validate: func(t *testing.T, token string, config *Config) {
				t.Helper()
				claims, err := ValidateToken(token, config)
				require.NoError(t, err)
				assert.Equal(t, "user456", claims.UserID)
			},
			wantErr:     nil,
			description: "should generate token with only user_id",
		},
		{
			name:        "invalid config",
			setupClaims: createTestClaims,
			setupConfig: func(t *testing.T) *Config {
				t.Helper()
				return &Config{
					Secret:          "",
					SigningMethod:   jwt.SigningMethodHS256,
					AccessDuration:  15 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
				}
			},
			validate:    nil,
			wantErr:     ErrSecretRequired,
			description: "should fail with invalid config",
		},
		{
			name:        "token with HS512 signing",
			setupClaims: createTestClaims,
			setupConfig: func(t *testing.T) *Config {
				t.Helper()
				return &Config{
					Secret:          "test-secret-key-for-hs512",
					SigningMethod:   jwt.SigningMethodHS512,
					AccessDuration:  15 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
					Issuer:          "govern-test",
				}
			},
			validate: func(t *testing.T, token string, config *Config) {
				t.Helper()
				assertTokenValid(t, token, config)
			},
			wantErr:     nil,
			description: "should generate token with HS512",
		},
		{
			name:        "token expiration set correctly",
			setupClaims: createTestClaims,
			setupConfig: func(t *testing.T) *Config {
				t.Helper()
				return &Config{
					Secret:          "test-secret",
					SigningMethod:   jwt.SigningMethodHS256,
					AccessDuration:  30 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
					Issuer:          "govern-test",
				}
			},
			validate: func(t *testing.T, token string, config *Config) {
				t.Helper()
				claims, err := ValidateToken(token, config)
				require.NoError(t, err)
				assert.NotNil(t, claims.ExpiresAt)
				expiryTime := claims.ExpiresAt.Time
				expectedMinExpiry := time.Now().Add(29 * time.Minute)
				expectedMaxExpiry := time.Now().Add(31 * time.Minute)
				assert.True(t, expiryTime.After(expectedMinExpiry) && expiryTime.Before(expectedMaxExpiry),
					"expiry should be approximately 30 minutes from now")
			},
			wantErr:     nil,
			description: "should set correct expiration time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := tt.setupClaims(t)
			config := tt.setupConfig(t)

			token, err := GenerateAccessToken(claims, config)

			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				if tt.validate != nil {
					tt.validate(t, token, config)
				}
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		setupClaims func(*testing.T) *Claims
		setupConfig func(*testing.T) *Config
		validate    func(*testing.T, string, *Config)
		wantErr     error
		description string
	}{
		{
			name:        "valid refresh token",
			setupClaims: createTestClaims,
			setupConfig: createTestConfig,
			validate: func(t *testing.T, token string, config *Config) {
				t.Helper()
				claims, err := ValidateToken(token, config)
				require.NoError(t, err)
				assert.Equal(t, "user123", claims.UserID)
				assert.Empty(t, claims.Username)
				assert.Empty(t, claims.Email)
			},
			wantErr:     nil,
			description: "should generate valid refresh token with minimal claims",
		},
		{
			name:        "invalid config for refresh token",
			setupClaims: createTestClaims,
			setupConfig: func(t *testing.T) *Config {
				t.Helper()
				return &Config{
					Secret:          "",
					SigningMethod:   jwt.SigningMethodHS256,
					AccessDuration:  15 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
				}
			},
			validate:    nil,
			wantErr:     ErrSecretRequired,
			description: "should fail with invalid config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := tt.setupClaims(t)
			config := tt.setupConfig(t)

			token, err := GenerateRefreshToken(claims, config)

			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				if tt.validate != nil {
					tt.validate(t, token, config)
				}
			}
		})
	}
}

// Token validation tests

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		setupToken  func(*testing.T) string
		setupConfig func(*testing.T) *Config
		wantErr     error
		validate    func(*testing.T, *Claims)
		description string
	}{
		{
			name: "valid token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				config := createTestConfig(t)
				claims := createTestClaims(t)
				token, err := GenerateAccessToken(claims, config)
				require.NoError(t, err)
				return token
			},
			setupConfig: createTestConfig,
			wantErr:     nil,
			validate: func(t *testing.T, claims *Claims) {
				t.Helper()
				assert.Equal(t, "user123", claims.UserID)
				assert.Equal(t, "testuser", claims.Username)
				assert.Equal(t, "test@example.com", claims.Email)
				assert.Contains(t, claims.Roles, "admin")
			},
			description: "should validate correct token",
		},
		{
			name: "invalid signature",
			setupToken: func(t *testing.T) string {
				t.Helper()
				config := createTestConfig(t)
				claims := createTestClaims(t)
				token, _ := GenerateAccessToken(claims, config)
				// Corrupt the token
				parts := strings.Split(token, ".")
				if len(parts) == 3 {
					parts[2] = "invalidsignature" + parts[2][:10]
					return strings.Join(parts, ".")
				}
				return token
			},
			setupConfig: createTestConfig,
			wantErr:     ErrSignatureInvalid,
			description: "should reject token with invalid signature",
		},
		{
			name: "malformed token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				return "invalid.token.format"
			},
			setupConfig: createTestConfig,
			wantErr:     ErrTokenMalformed,
			description: "should reject malformed token",
		},
		{
			name: "empty token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			setupConfig: createTestConfig,
			wantErr:     ErrTokenMalformed,
			description: "should reject empty token",
		},
		{
			name: "expired token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				// Create a token with past expiration
				claims := createTestClaims(t)
				claims.RegisteredClaims = jwt.RegisteredClaims{
					Issuer:    "govern-test",
					Subject:   claims.UserID,
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte("test-secret-key-for-testing-purposes-only"))
				require.NoError(t, err)
				return tokenString
			},
			setupConfig: createTestConfig,
			wantErr:     ErrTokenExpired,
			description: "should reject expired token",
		},
		{
			name: "token with wrong secret",
			setupToken: func(t *testing.T) string {
				t.Helper()
				config := &Config{
					Secret:          "different-secret",
					SigningMethod:   jwt.SigningMethodHS256,
					AccessDuration:  15 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
					Issuer:          "govern-test",
				}
				claims := createTestClaims(t)
				token, err := GenerateAccessToken(claims, config)
				require.NoError(t, err)
				return token
			},
			setupConfig: createTestConfig,
			wantErr:     ErrSignatureInvalid,
			description: "should reject token signed with different secret",
		},
		{
			name: "token with unexpected signing method",
			setupToken: func(t *testing.T) string {
				t.Helper()
				claims := createTestClaims(t)
				token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
				tokenString, err := token.SignedString([]byte("secret"))
				require.NoError(t, err)
				return tokenString
			},
			setupConfig: func(t *testing.T) *Config {
				t.Helper()
				return &Config{
					Secret:          "secret",
					SigningMethod:   jwt.SigningMethodHS256,
					AccessDuration:  15 * time.Minute,
					RefreshDuration: 7 * 24 * time.Hour,
					Issuer:          "govern-test",
				}
			},
			wantErr:     goverrors.NewCode(goverrors.CodeUnauthorized, "invalid authorization token"),
			description: "should reject token with unexpected signing method",
		},
		{
			name:        "invalid config",
			setupToken:  func(t *testing.T) string { return "any.token" },
			setupConfig: func(t *testing.T) *Config { return &Config{} },
			wantErr:     ErrSecretRequired,
			description: "should fail validation with invalid config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken(t)
			config := tt.setupConfig(t)

			claims, err := ValidateToken(token, config)

			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, claims)
				}
			}
		})
	}
}

func TestRefreshAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		setupToken  func(*testing.T) string
		setupConfig func(*testing.T) *Config
		validate    func(*testing.T, string, *Config)
		wantErr     error
		description string
	}{
		{
			name: "valid refresh token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				config := createTestConfig(t)
				claims := createTestClaims(t)
				token, err := GenerateRefreshToken(claims, config)
				require.NoError(t, err)
				return token
			},
			setupConfig: createTestConfig,
			validate: func(t *testing.T, newToken string, config *Config) {
				t.Helper()
				assert.NotEmpty(t, newToken)
				claims, err := ValidateToken(newToken, config)
				require.NoError(t, err)
				assert.Equal(t, "user123", claims.UserID)
			},
			wantErr:     nil,
			description: "should generate new access token from refresh token",
		},
		{
			name: "expired refresh token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				// Create a token with past expiration
				claims := createTestClaims(t)
				claims.RegisteredClaims = jwt.RegisteredClaims{
					Issuer:    "govern-test",
					Subject:   claims.UserID,
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte("test-secret-key-for-testing-purposes-only"))
				require.NoError(t, err)
				return tokenString
			},
			setupConfig: createTestConfig,
			wantErr:     ErrTokenExpired,
			description: "should reject expired refresh token",
		},
		{
			name: "invalid refresh token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				return "invalid.refresh.token"
			},
			setupConfig: createTestConfig,
			wantErr:     ErrTokenMalformed,
			description: "should reject invalid refresh token",
		},
		{
			name: "access token instead of refresh token",
			setupToken: func(t *testing.T) string {
				t.Helper()
				config := createTestConfig(t)
				claims := createTestClaims(t)
				token, err := GenerateAccessToken(claims, config)
				require.NoError(t, err)
				return token
			},
			setupConfig: createTestConfig,
			validate: func(t *testing.T, newToken string, config *Config) {
				t.Helper()
				// Access token can also be used to refresh (flexible design)
				assert.NotEmpty(t, newToken)
			},
			wantErr:     nil,
			description: "should accept access token for refresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := tt.setupToken(t)
			config := tt.setupConfig(t)

			newToken, err := RefreshAccessToken(refreshToken, config)

			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, newToken, config)
				}
			}
		})
	}
}

// Claims tests

func TestClaims_Validate(t *testing.T) {
	tests := []struct {
		name        string
		claims      *Claims
		wantErr     bool
		errContains string
		description string
	}{
		{
			name: "valid claims",
			claims: &Claims{
				UserID: "user123",
			},
			wantErr:     false,
			description: "should pass validation with user_id",
		},
		{
			name:        "missing user_id",
			claims:      &Claims{},
			wantErr:     true,
			errContains: "user_id is required",
			description: "should fail validation without user_id",
		},
		{
			name: "claims with all fields",
			claims: &Claims{
				UserID:   "user123",
				Username: "testuser",
				Email:    "test@example.com",
				Roles:    []string{"admin", "user"},
			},
			wantErr:     false,
			description: "should pass validation with all fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name        string
		claims      *Claims
		role        string
		want        bool
		description string
	}{
		{
			name: "has role",
			claims: &Claims{
				Roles: []string{"admin", "user"},
			},
			role:        "admin",
			want:        true,
			description: "should return true for existing role",
		},
		{
			name: "does not have role",
			claims: &Claims{
				Roles: []string{"user", "guest"},
			},
			role:        "admin",
			want:        false,
			description: "should return false for non-existing role",
		},
		{
			name:        "empty roles",
			claims:      &Claims{Roles: []string{}},
			role:        "admin",
			want:        false,
			description: "should return false with no roles",
		},
		{
			name:        "nil roles",
			claims:      &Claims{},
			role:        "admin",
			want:        false,
			description: "should return false with nil roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.HasRole(tt.role)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClaims_HasAnyRole(t *testing.T) {
	tests := []struct {
		name        string
		claims      *Claims
		roles       []string
		want        bool
		description string
	}{
		{
			name: "has one of the roles",
			claims: &Claims{
				Roles: []string{"admin", "user"},
			},
			roles:       []string{"admin", "moderator"},
			want:        true,
			description: "should return true if has any role",
		},
		{
			name: "has multiple roles",
			claims: &Claims{
				Roles: []string{"admin", "user", "moderator"},
			},
			roles:       []string{"guest", "admin"},
			want:        true,
			description: "should return true if matches any role",
		},
		{
			name: "has none of the roles",
			claims: &Claims{
				Roles: []string{"user", "guest"},
			},
			roles:       []string{"admin", "moderator"},
			want:        false,
			description: "should return false if has none of the roles",
		},
		{
			name:        "empty role list",
			claims:      &Claims{Roles: []string{"admin"}},
			roles:       []string{},
			want:        false,
			description: "should return false with empty role list",
		},
		{
			name:        "nil roles in claims",
			claims:      &Claims{},
			roles:       []string{"admin"},
			want:        false,
			description: "should return false with nil claims roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.HasAnyRole(tt.roles...)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Secure secret generation tests

func TestGenerateSecureSecret(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		validate    func(*testing.T, string)
		wantErr     bool
		description string
	}{
		{
			name:   "generate 32 byte secret",
			length: 32,
			validate: func(t *testing.T, secret string) {
				t.Helper()
				assert.NotEmpty(t, secret)
				// Base64 encoding: 32 bytes -> 44 characters
				assert.Equal(t, 44, len(secret))
			},
			wantErr:     false,
			description: "should generate 32-byte secret",
		},
		{
			name:   "generate 64 byte secret",
			length: 64,
			validate: func(t *testing.T, secret string) {
				t.Helper()
				assert.NotEmpty(t, secret)
				assert.Equal(t, 88, len(secret))
			},
			wantErr:     false,
			description: "should generate 64-byte secret",
		},
		{
			name:   "generate 16 byte secret",
			length: 16,
			validate: func(t *testing.T, secret string) {
				t.Helper()
				assert.NotEmpty(t, secret)
				assert.Equal(t, 24, len(secret))
			},
			wantErr:     false,
			description: "should generate 16-byte secret",
		},
		{
			name:        "zero length",
			length:      0,
			validate:    nil,
			wantErr:     false,
			description: "should handle zero length",
		},
		{
			name:   "unique secrets",
			length: 32,
			validate: func(t *testing.T, secret string) {
				t.Helper()
				// Generate another secret and ensure they're different
				secret2, err := GenerateSecureSecret(32)
				require.NoError(t, err)
				assert.NotEqual(t, secret, secret2, "secrets should be unique")
			},
			wantErr:     false,
			description: "should generate unique secrets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := GenerateSecureSecret(tt.length)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, secret)
				}
			}
		})
	}
}

// Middleware tests

func TestDefaultTokenExtractor(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func(*testing.T) *http.Request
		wantErr     error
		validate    func(*testing.T, string)
		description string
	}{
		{
			name: "valid bearer token",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				req.Header.Set("Authorization", "Bearer valid-token-123")
				return req
			},
			wantErr: nil,
			validate: func(t *testing.T, token string) {
				t.Helper()
				assert.Equal(t, "valid-token-123", token)
			},
			description: "should extract valid bearer token",
		},
		{
			name: "missing authorization header",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest("GET", "/test", http.NoBody)
			},
			wantErr:     ErrTokenMissing,
			description: "should return error for missing header",
		},
		{
			name: "invalid authorization format - no bearer",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				req.Header.Set("Authorization", "valid-token-123")
				return req
			},
			wantErr:     ErrTokenInvalid,
			description: "should return error for invalid format",
		},
		{
			name: "invalid authorization format - wrong prefix",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				req.Header.Set("Authorization", "Basic token-123")
				return req
			},
			wantErr:     ErrTokenInvalid,
			description: "should return error for wrong prefix",
		},
		{
			name: "empty bearer token",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				req.Header.Set("Authorization", "Bearer ")
				return req
			},
			wantErr:     ErrTokenMissing,
			description: "should return error for empty token",
		},
		{
			name: "multiple spaces in bearer token",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				req.Header.Set("Authorization", "Bearer   token-with-spaces")
				return req
			},
			wantErr: nil,
			validate: func(t *testing.T, token string) {
				t.Helper()
				assert.Equal(t, "  token-with-spaces", token)
			},
			description: "should handle multiple spaces (keeps after Bearer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq(t)

			token, err := DefaultTokenExtractor(req)

			if tt.wantErr != nil {
				assertErrorCode(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, token)
				}
			}
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func(*testing.T) *http.Request
		validate    func(*testing.T, *Claims, bool)
		description string
	}{
		{
			name: "user in context",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				claims := createTestClaims(t)
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				ctx := req.Context()
				ctx = context.WithValue(ctx, UserKey, claims)
				return req.WithContext(ctx)
			},
			validate: func(t *testing.T, claims *Claims, ok bool) {
				t.Helper()
				assert.True(t, ok)
				assert.NotNil(t, claims)
				assert.Equal(t, "user123", claims.UserID)
			},
			description: "should retrieve user from context",
		},
		{
			name: "no user in context",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest("GET", "/test", http.NoBody)
			},
			validate: func(t *testing.T, claims *Claims, ok bool) {
				t.Helper()
				assert.False(t, ok)
				assert.Nil(t, claims)
			},
			description: "should return false when no user in context",
		},
		{
			name: "wrong type in context",
			setupReq: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/test", http.NoBody)
				ctx := req.Context()
				ctx = context.WithValue(ctx, UserKey, "not-a-claims")
				return req.WithContext(ctx)
			},
			validate: func(t *testing.T, claims *Claims, ok bool) {
				t.Helper()
				assert.False(t, ok)
				assert.Nil(t, claims)
			},
			description: "should handle wrong type in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq(t)
			claims, ok := GetCurrentUser(req)
			tt.validate(t, claims, ok)
		})
	}
}

func TestMustGetCurrentUser(t *testing.T) {
	t.Run("should panic when no user in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		assert.Panics(t, func() {
			MustGetCurrentUser(req)
		})
	})

	t.Run("should return user when in context", func(t *testing.T) {
		claims := createTestClaims(t)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		ctx := req.Context()
		ctx = context.WithValue(ctx, UserKey, claims)
		req = req.WithContext(ctx)

		result := MustGetCurrentUser(req)
		assert.Equal(t, claims, result)
	})
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*testing.T) *MiddlewareConfig
		setupReq    func(*testing.T, string) *http.Request
		validate    func(*testing.T, *httptest.ResponseRecorder)
		description string
	}{
		{
			name: "valid token passes through",
			setupConfig: func(t *testing.T) *MiddlewareConfig {
				t.Helper()
				config := createTestConfig(t)
				return &MiddlewareConfig{
					Config:         config,
					TokenExtractor: DefaultTokenExtractor,
					ErrorHandler:   DefaultErrorHandler,
				}
			},
			setupReq: func(t *testing.T, token string) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/protected", http.NoBody)
				req.Header.Set("Authorization", "Bearer "+token)
				return req
			},
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusOK, w.Code)
				// Handler includes user info in response
				assert.Contains(t, w.Body.String(), "protected")
			},
			description: "should allow request with valid token",
		},
		{
			name: "missing token returns 401",
			setupConfig: func(t *testing.T) *MiddlewareConfig {
				t.Helper()
				config := createTestConfig(t)
				return &MiddlewareConfig{
					Config:         config,
					TokenExtractor: DefaultTokenExtractor,
					ErrorHandler:   DefaultErrorHandler,
				}
			},
			setupReq: func(t *testing.T, token string) *http.Request {
				t.Helper()
				return httptest.NewRequest("GET", "/protected", http.NoBody)
			},
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, w.Code)
			},
			description: "should reject request without token",
		},
		{
			name: "invalid token returns 401",
			setupConfig: func(t *testing.T) *MiddlewareConfig {
				t.Helper()
				config := createTestConfig(t)
				return &MiddlewareConfig{
					Config:         config,
					TokenExtractor: DefaultTokenExtractor,
					ErrorHandler:   DefaultErrorHandler,
				}
			},
			setupReq: func(t *testing.T, token string) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/protected", http.NoBody)
				req.Header.Set("Authorization", "Bearer invalid-token")
				return req
			},
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, w.Code)
			},
			description: "should reject request with invalid token",
		},
		{
			name: "skip path bypasses validation",
			setupConfig: func(t *testing.T) *MiddlewareConfig {
				t.Helper()
				config := createTestConfig(t)
				return &MiddlewareConfig{
					Config:         config,
					TokenExtractor: DefaultTokenExtractor,
					ErrorHandler:   DefaultErrorHandler,
					SkipPaths:      []string{"/public"},
				}
			},
			setupReq: func(t *testing.T, token string) *http.Request {
				t.Helper()
				return httptest.NewRequest("GET", "/public/resource", http.NoBody)
			},
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "public", w.Body.String())
			},
			description: "should skip authentication for skip paths",
		},
		{
			name: "context contains user claims",
			setupConfig: func(t *testing.T) *MiddlewareConfig {
				t.Helper()
				config := createTestConfig(t)
				return &MiddlewareConfig{
					Config:         config,
					TokenExtractor: DefaultTokenExtractor,
					ErrorHandler:   DefaultErrorHandler,
				}
			},
			setupReq: func(t *testing.T, token string) *http.Request {
				t.Helper()
				req := httptest.NewRequest("GET", "/protected", http.NoBody)
				req.Header.Set("Authorization", "Bearer "+token)
				return req
			},
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusOK, w.Code)
				// Handler checks for user in context
				assert.Contains(t, w.Body.String(), "user123")
			},
			description: "should add user to request context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig(t)

			// Generate a valid token for protected routes
			var validToken string
			if strings.Contains(tt.name, "valid token") || strings.Contains(tt.name, "context") {
				claims := createTestClaims(t)
				var err error
				validToken, err = GenerateAccessToken(claims, config.Config)
				require.NoError(t, err)
			}

			// Create handler that writes response
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/public") {
					_, _ = w.Write([]byte("public"))
					return
				}

				claims, ok := GetCurrentUser(r)
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if claims.UserID == "user123" {
					_, _ = w.Write([]byte("protected with user: " + claims.UserID))
				} else {
					_, _ = w.Write([]byte("protected"))
				}
			})

			middleware := Middleware(config)
			req := tt.setupReq(t, validToken)
			w := httptest.NewRecorder()

			middleware(handler).ServeHTTP(w, req)
			tt.validate(t, w)
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		skipPaths   []string
		want        bool
		description string
	}{
		{
			name:        "exact match",
			path:        "/public",
			skipPaths:   []string{"/public"},
			want:        true,
			description: "should skip exact path match",
		},
		{
			name:        "prefix match",
			path:        "/public/resource",
			skipPaths:   []string{"/public"},
			want:        true,
			description: "should skip path with prefix",
		},
		{
			name:        "no match",
			path:        "/protected",
			skipPaths:   []string{"/public"},
			want:        false,
			description: "should not skip different path",
		},
		{
			name:        "multiple skip paths",
			path:        "/health",
			skipPaths:   []string{"/public", "/health"},
			want:        true,
			description: "should skip when path matches any in list",
		},
		{
			name:        "empty skip paths",
			path:        "/public",
			skipPaths:   []string{},
			want:        false,
			description: "should not skip with empty list",
		},
		{
			name:        "root path",
			path:        "/",
			skipPaths:   []string{"/"},
			want:        true,
			description: "should skip root path",
		},
		{
			name:        "nested prefix",
			path:        "/api/v1/health",
			skipPaths:   []string{"/api/v1"},
			want:        true,
			description: "should skip nested path prefix",
		},
		{
			name:        "case sensitive",
			path:        "/Public",
			skipPaths:   []string{"/public"},
			want:        false,
			description: "should be case sensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipPath(tt.path, tt.skipPaths)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Error handling tests

func TestHandleJWTError(t *testing.T) {
	tests := []struct {
		name        string
		inputErr    error
		wantErr     error
		description string
	}{
		{
			name:        "nil error",
			inputErr:    nil,
			wantErr:     nil,
			description: "should return nil for nil input",
		},
		{
			name:        "expired token error",
			inputErr:    stderrors.New("token is expired"),
			wantErr:     ErrTokenExpired,
			description: "should convert expired error",
		},
		{
			name:        "signature invalid error",
			inputErr:    stderrors.New("signature is invalid"),
			wantErr:     ErrSignatureInvalid,
			description: "should convert signature error",
		},
		{
			name:        "malformed token error",
			inputErr:    stderrors.New("token contains an invalid number of segments"),
			wantErr:     ErrTokenMalformed,
			description: "should convert malformed error",
		},
		{
			name:        "unknown error",
			inputErr:    stderrors.New("unknown jwt error"),
			wantErr:     goverrors.NewCode(goverrors.CodeUnauthorized, "invalid authorization token"),
			description: "should wrap unknown errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleJWTError(tt.inputErr)

			if tt.wantErr == nil {
				assert.Nil(t, got)
			} else {
				assert.Error(t, got)
				// Check error code matches
				code1, ok1 := goverrors.GetCode(tt.wantErr)
				code2, ok2 := goverrors.GetCode(got)
				if ok1 && ok2 {
					assert.Equal(t, code1, code2)
				}
			}
		})
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		validate    func(*testing.T, *httptest.ResponseRecorder)
		description string
	}{
		{
			name: "unauthorized error",
			err:  ErrTokenInvalid,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
			},
			description: "should return 401 with WWW-Authenticate header",
		},
		{
			name: "token expired error",
			err:  ErrTokenExpired,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
			},
			description: "should return 401 for expired token",
		},
		{
			name: "other error",
			err:  stderrors.New("some other error"),
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusUnauthorized, w.Code)
			},
			description: "should return 401 for other errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			DefaultErrorHandler(w, req, tt.err)
			tt.validate(t, w)
		})
	}
}
