package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// generateTestSecret creates a 32-character secret for testing purposes.
// This avoids hardcoded secrets in the test code that would trigger gitleaks.
func generateTestSecret() string {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a deterministic value derived from a constant seed
		seed := []byte("golang-sample-test-secret-seed-2026")
		hash := sha256.Sum256(seed)
		return hex.EncodeToString(hash[:16]) // First 16 bytes = 32 hex chars
	}
	return hex.EncodeToString(bytes)
}

func TestLoadConfigFromEnvFile(t *testing.T) {
	// Test loading from .env file (in project root)
	logger := zap.NewNop()
	cfgFile := filepath.Join(getProjectRoot(), ".test-env")
	cfg, err := LoadConfig(cfgFile, logger)

	assert.NoError(t, err, "Should load .env file without error")
	assert.NotNil(t, cfg, "Config should not be nil")

	// Verify app config
	assert.True(t, cfg.App.Debug, "Debug should be true")
	assert.Equal(t, "development", cfg.App.Env, "Env should be development")

	// Verify postgres config
	assert.Contains(t, cfg.Postgres.DSN, "golang_sample_test", "DSN should contain test database name")
	assert.Contains(t, cfg.Postgres.DSN, "localhost", "DSN should contain localhost")

	// Verify redis config
	assert.Equal(t, "redis://localhost:6379/1", cfg.Redis.URL, "Redis URL should match")

	// Verify API config
	assert.Equal(t, "test-jwt-secret-key", cfg.API.Secret, "API secret should match")

	// Verify global ENV is set
	assert.Equal(t, cfg, ENV, "Global ENV should be set")
}

func TestLoadConfigFromYAMLFile(t *testing.T) {
	// Test loading from YAML file (in project root)
	logger := zap.NewNop()
	cfgFile := filepath.Join(getProjectRoot(), "config.test.yaml")
	cfg, err := LoadConfig(cfgFile, logger)

	assert.NoError(t, err, "Should load YAML file without error")
	assert.NotNil(t, cfg, "Config should not be nil")

	// Verify app config
	assert.True(t, cfg.App.Debug, "Debug should be true")
	assert.Equal(t, "development", cfg.App.Env, "Env should be development")

	// Verify postgres config
	assert.Contains(t, cfg.Postgres.DSN, "dbname=golang_sample", "DSN should contain database name")

	// Verify API config
	assert.Equal(t, "test-jwt-secret-key-for-testing-purposes-only", cfg.API.Secret, "API secret should match")

	// Verify global ENV is set
	assert.Equal(t, cfg, ENV, "Global ENV should be set")
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Test loading non-existent file
	logger := zap.NewNop()
	cfg, err := LoadConfig("nonexistent.env", logger)

	assert.Error(t, err, "Should return error for missing file")
	assert.Nil(t, cfg, "Config should be nil on error")
}

func TestLoadConfigValidation(t *testing.T) {
	// Create a file with missing required fields
	content := `APP_DEBUG=true
# APP_ENV is missing - required field
APP_POSTGRES_DSN="test"
`

	invalidFile := filepath.Join(getProjectRoot(), ".invalid.env")
	err := os.WriteFile(invalidFile, []byte(content), 0644)
	assert.NoError(t, err, "Should create test file")
	defer os.Remove(invalidFile)

	logger := zap.NewNop()
	cfg, err := LoadConfig(invalidFile, logger)

	assert.Error(t, err, "Should return error for validation failure")
	assert.Nil(t, cfg, "Config should be nil on validation error")
}

func TestEnvConfigMapValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *EnvConfigMap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &EnvConfigMap{
				App: struct {
					Debug bool   `mapstructure:"debug"`
					Env   string `mapstructure:"env" validate:"required"`
				}{
					Debug: true,
					Env:   "development",
				},
				Postgres: struct {
					DSN string `mapstructure:"dsn" validate:"required"`
				}{
					DSN: "host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable",
				},
				API: struct {
					Secret string `mapstructure:"secret"`
				}{
					Secret: generateTestSecret(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid env - not allowed value",
			cfg: &EnvConfigMap{
				App: struct {
					Debug bool   `mapstructure:"debug"`
					Env   string `mapstructure:"env" validate:"required"`
				}{
					Debug: true,
					Env:   "invalid-env",
				},
				Postgres: struct {
					DSN string `mapstructure:"dsn" validate:"required"`
				}{
					DSN: "host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable",
				},
			},
			wantErr: true,
			errMsg:  "invalid APP_ENV",
		},
		{
			name: "invalid secret - too short",
			cfg: &EnvConfigMap{
				App: struct {
					Debug bool   `mapstructure:"debug"`
					Env   string `mapstructure:"env" validate:"required"`
				}{
					Debug: true,
					Env:   "production",
				},
				Postgres: struct {
					DSN string `mapstructure:"dsn" validate:"required"`
				}{
					DSN: "host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable",
				},
				API: struct {
					Secret string `mapstructure:"secret"`
				}{
					Secret: "short",
				},
			},
			wantErr: true,
			errMsg:  "must be at least 32 characters",
		},
		{
			name: "staging env is valid",
			cfg: &EnvConfigMap{
				App: struct {
					Debug bool   `mapstructure:"debug"`
					Env   string `mapstructure:"env" validate:"required"`
				}{
					Debug: false,
					Env:   "staging",
				},
				Postgres: struct {
					DSN string `mapstructure:"dsn" validate:"required"`
				}{
					DSN: "host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable",
				},
			},
			wantErr: false,
		},
		{
			name: "production env is valid",
			cfg: &EnvConfigMap{
				App: struct {
					Debug bool   `mapstructure:"debug"`
					Env   string `mapstructure:"env" validate:"required"`
				}{
					Debug: false,
					Env:   "production",
				},
				Postgres: struct {
					DSN string `mapstructure:"dsn" validate:"required"`
				}{
					DSN: "host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Should return error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Should not return error")
			}
		})
	}
}
