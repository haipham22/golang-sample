package rest

import (
	"testing"

	"github.com/haipham22/golang-sample/pkg/config"
	"go.uber.org/zap"
)

// TestNew_MissingJWTSecret verifies the config guard fails before any DB
// connection is attempted. This is the unit-testable path; the success path
// requires a live database and is covered by integration/handler tests.
func TestNew_MissingJWTSecret(t *testing.T) {
	cfg := &config.EnvConfigMap{} // API.Secret empty

	_, cleanup, err := New(zap.NewNop().Sugar(), 8080, cfg)
	if err == nil {
		if cleanup != nil {
			cleanup()
		}
		t.Fatal("expected error for missing JWT secret, got nil")
	}
	if err != ErrMissingJWTSecret {
		t.Errorf("err = %v, want ErrMissingJWTSecret", err)
	}
}

// TestNew_DBErrorOnBadDSN verifies a bad DSN surfaces a database error (still
// no panic). Uses an invalid DSN so connection fails fast.
func TestNew_DBErrorOnBadDSN(t *testing.T) {
	cfg := &config.EnvConfigMap{}
	cfg.API.Secret = "test-secret-at-least-32-characters-long"
	cfg.Postgres.DSN = "invalid-dsn://not-valid"

	_, _, err := New(zap.NewNop().Sugar(), 8080, cfg)
	if err == nil {
		t.Fatal("expected DB error for invalid DSN, got nil")
	}
}
