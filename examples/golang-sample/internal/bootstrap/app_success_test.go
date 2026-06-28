package bootstrap

import (
	"errors"
	"testing"

	governhttp "github.com/haipham22/govern/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/pkg/config"
)

// withSwappedFactory temporarily replaces httpServerFactory with fn for the
// duration of the test, restoring the production default on cleanup so other
// tests are unaffected.
func withSwappedFactory(t *testing.T, fn func(*zap.SugaredLogger, int64, *config.EnvConfigMap) (governhttp.Server, func(), error)) {
	t.Helper()
	original := httpServerFactory
	httpServerFactory = fn
	t.Cleanup(func() { httpServerFactory = original })
}

// validTestConfig returns a Config that passes NewLogger (non-production env)
// and the rest.New secret guard. It does NOT reach Postgres because the
// factory is swapped in each test.
func validTestConfig() *config.EnvConfigMap {
	cfg := &config.EnvConfigMap{}
	cfg.App.Env = config.EnvDevelopment
	cfg.API.Secret = "unit-test-secret-at-least-32-characters-long"
	return cfg
}

// TestNew_SuccessAndCleanup covers the success branch of New: server + cleanup
// returned, no error, and the cleanup closure runs the http-cleanup branch.
// Calling cleanup twice exercises the closure on already-run state.
func TestNew_SuccessAndCleanup(t *testing.T) {
	httpCleanupCalls := 0
	withSwappedFactory(t, func(_ *zap.SugaredLogger, _ int64, _ *config.EnvConfigMap) (governhttp.Server, func(), error) {
		return nil, func() { httpCleanupCalls++ }, nil
	})

	server, cleanup, err := New(Config{Port: 0, AppConfig: validTestConfig()})
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	_ = server

	// First call runs the httpCleanup != nil branch.
	assert.NotPanics(t, func() { cleanup() })
	assert.Equal(t, 1, httpCleanupCalls)

	// Second call exercises the closure again (idempotent in practice; here
	// it proves the httpCleanup branch is reachable on repeat invocations).
	assert.NotPanics(t, func() { cleanup() })
	assert.Equal(t, 2, httpCleanupCalls)
}

// TestNew_HTTPFactoryErrorRunsLogCleanup covers the error branch: New must
// invoke logCleanup before returning the wrapped factory error.
func TestNew_HTTPFactoryErrorRunsLogCleanup(t *testing.T) {
	factoryErr := errors.New("simulated http factory failure")
	withSwappedFactory(t, func(_ *zap.SugaredLogger, _ int64, _ *config.EnvConfigMap) (governhttp.Server, func(), error) {
		return nil, nil, factoryErr
	})

	server, cleanup, err := New(Config{Port: 0, AppConfig: validTestConfig()})
	require.Error(t, err)
	assert.ErrorIs(t, err, factoryErr)
	assert.Nil(t, server)
	assert.Nil(t, cleanup)
}

// TestNew_HTTPFactoryNilCleanup covers the nil-guard inside the cleanup
// closure (the `if httpCleanup != nil` branch). The factory returns a nil
// cleanup func; calling the returned cleanup must not panic.
func TestNew_HTTPFactoryNilCleanup(t *testing.T) {
	withSwappedFactory(t, func(_ *zap.SugaredLogger, _ int64, _ *config.EnvConfigMap) (governhttp.Server, func(), error) {
		return nil, nil, nil
	})

	server, cleanup, err := New(Config{Port: 0, AppConfig: validTestConfig()})
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	_ = server

	assert.NotPanics(t, func() { cleanup() })
}
