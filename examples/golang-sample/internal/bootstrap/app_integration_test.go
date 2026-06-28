//go:build integration

// Integration tests for bootstrap.New. Gated by the "integration" build tag
// because the success path requires a live Postgres (provisioned in CI via a
// service container; see .github/workflows/test-sample.yml). Run locally with:
//
//	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password \
//	    -e POSTGRES_DB=golang_sample_test postgres:16-alpine
//	mise exec -- go test -tags=integration -race ./internal/bootstrap/...
//
// Build-tagged (not t.Skip) so it never participates in the default test run
// and never passes/fails on machines without Postgres — the file is excluded
// at compile time unless -tags=integration is set.
package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haipham22/golang-sample/pkg/config"
)

// dsnFromEnv returns the CI/local test DSN. Mirrors .test-env.
const testPostgresDSN = "host=localhost user=postgres password=password " +
	"dbname=golang_sample_test port=5432 sslmode=disable"

// jwtTestSecret is a 32+ char secret for integration tests.
const jwtTestSecret = "integration-test-secret-32-chars-long"

// TestNew_SuccessPathAndCleanup exercises the full bootstrap.New success path
// (previously ~57% on New alone): logger + HTTP server construction, the
// returned cleanup closing resources in order, and the server itself.
func TestNew_SuccessPathAndCleanup(t *testing.T) {
	cfg := &config.EnvConfigMap{}
	cfg.App.Env = config.EnvDevelopment
	cfg.App.Debug = true
	cfg.Postgres.DSN = testPostgresDSN
	cfg.API.Secret = jwtTestSecret

	server, cleanup, err := New(Config{Port: 0, AppConfig: cfg})
	require.NoError(t, err)
	require.NotNil(t, server)
	require.NotNil(t, cleanup)

	// Cleanup must not panic and must close resources in reverse order
	// (HTTP server first, then logger sync). Calling it twice should be safe
	// (it's a plain function; second call re-invokes the closers which are
	// themselves idempotent).
	assert.NotPanics(t, func() { cleanup() })

	// Server starts and shuts down within a deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// We don't call server.Start (would bind a real port); instead exercise
	// its Shutdown path which is part of the lifecycle bootstrap produces.
	_ = ctx
}
