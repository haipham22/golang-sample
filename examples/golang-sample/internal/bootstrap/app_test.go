package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haipham22/golang-sample/pkg/config"
)

// TestNew_NilConfig verifies the bootstrap guard rejects a nil config before
// touching the logger or HTTP stack.
func TestNew_NilConfig(t *testing.T) {
	_, cleanup, err := New(Config{AppConfig: nil})
	require.Error(t, err)
	assert.Nil(t, cleanup)
	assert.Contains(t, err.Error(), "nil app config")
}

// TestNew_MissingJWTSecret verifies the guard propagates from rest.New when
// the API secret is empty (no DB connection attempted).
func TestNew_MissingJWTSecret(t *testing.T) {
	cfg := &config.EnvConfigMap{} // API.Secret empty

	_, cleanup, err := New(Config{Port: 8080, AppConfig: cfg})
	require.Error(t, err)
	assert.Nil(t, cleanup)
	// The wrapped error must mention the missing secret so operators know what
	// to fix.
	assert.Contains(t, err.Error(), "JWT secret")
}

// TestNewLogger_BuildsPerEnv verifies the logger factory returns a non-nil
// SugaredLogger for each supported environment.
func TestNewLogger_BuildsPerEnv(t *testing.T) {
	for _, env := range []string{config.EnvDevelopment, config.EnvStaging, config.EnvProduction} {
		t.Run(env, func(t *testing.T) {
			cfg := &config.EnvConfigMap{}
			cfg.App.Env = env

			log, cleanup, err := NewLogger(cfg)
			require.NoError(t, err)
			require.NotNil(t, log)
			require.NotNil(t, cleanup)
			assert.NotPanics(t, func() { cleanup() })
		})
	}
}

func TestNewLogger_NilConfig(t *testing.T) {
	_, _, err := NewLogger(nil)
	require.Error(t, err)
}
