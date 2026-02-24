package config

import (
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
	assert.Contains(t, cfg.Postgres.DSN, "golang_sample_yaml", "DSN should contain yaml database name")

	// Verify redis config
	assert.Equal(t, "redis://localhost:6379/2", cfg.Redis.URL, "Redis URL should match")

	// Verify API config
	assert.Equal(t, "yaml-jwt-secret-key", cfg.API.Secret, "API secret should match")

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
