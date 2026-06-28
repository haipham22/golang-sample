package rest

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/pkg/config"
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

func TestNew_SuccessWithInjectedDB(t *testing.T) {
	oldNewGormDB := newGormDB
	t.Cleanup(func() { newGormDB = oldNewGormDB })

	newGormDB = func(_ string, _ bool) (*gorm.DB, func(), error) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, nil, err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, nil, err
		}
		return db, func() { _ = sqlDB.Close() }, nil
	}

	server, cleanup, err := New(zap.NewNop().Sugar(), 0, validDIConfig(config.EnvDevelopment))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server == nil {
		t.Fatal("server is nil")
	}
	if cleanup == nil {
		t.Fatal("cleanup is nil")
	}
	cleanup()
}

func TestNew_AutoMigrateErrorCleansUp(t *testing.T) {
	oldNewGormDB := newGormDB
	t.Cleanup(func() { newGormDB = oldNewGormDB })

	cleanupCalled := false
	newGormDB = func(_ string, _ bool) (*gorm.DB, func(), error) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, nil, err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, nil, err
		}
		_ = sqlDB.Close()
		return db, func() { cleanupCalled = true }, nil
	}

	_, _, err := New(zap.NewNop().Sugar(), 0, validDIConfig(config.EnvDevelopment))
	if err == nil {
		t.Fatal("expected auto-migrate error, got nil")
	}
	if !cleanupCalled {
		t.Fatal("cleanup not called on auto-migrate error")
	}
}

func TestNew_ProductionSkipsAutoMigrate(t *testing.T) {
	oldNewGormDB := newGormDB
	t.Cleanup(func() { newGormDB = oldNewGormDB })

	newGormDB = func(_ string, _ bool) (*gorm.DB, func(), error) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, nil, err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, nil, err
		}
		_ = sqlDB.Close()
		return db, func() {}, nil
	}

	server, cleanup, err := New(zap.NewNop().Sugar(), 0, validDIConfig(config.EnvProduction))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if server == nil {
		t.Fatal("server is nil")
	}
	if cleanup == nil {
		t.Fatal("cleanup is nil")
	}
}

func TestNew_DBFactoryError(t *testing.T) {
	oldNewGormDB := newGormDB
	t.Cleanup(func() { newGormDB = oldNewGormDB })

	wantErr := errors.New("db factory failed")
	newGormDB = func(_ string, _ bool) (*gorm.DB, func(), error) {
		return nil, nil, wantErr
	}

	_, _, err := New(zap.NewNop().Sugar(), 0, validDIConfig(config.EnvDevelopment))
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func validDIConfig(env string) *config.EnvConfigMap {
	cfg := &config.EnvConfigMap{}
	cfg.App.Env = env
	cfg.API.Secret = "test-secret-at-least-32-characters-long"
	cfg.Postgres.DSN = "unused"
	return cfg
}
