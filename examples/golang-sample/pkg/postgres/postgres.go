package postgres

import (
	"time"

	governpostgres "github.com/haipham22/govern/database/postgres"
	"gorm.io/gorm"
)

// Config holds database configuration
type Config struct {
	DSN          string
	Debug        bool
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

// NewGormDB creates a new gorm postgresql with connection pooling
func NewGormDB(
	pgDSN string,
) (*gorm.DB, func(), error) {
	return New(Config{
		DSN:          pgDSN,
		Debug:        true,
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
		MaxIdleTime:  10 * time.Minute,
	})
}

// New creates a new gorm database with govern/postgres
func New(cfg Config) (*gorm.DB, func(), error) {
	// Build govern postgres options
	options := []governpostgres.Option{
		governpostgres.WithDebug(cfg.Debug),
		governpostgres.WithPreferSimpleProtocol(true),
	}

	// Add connection pooling options
	if cfg.MaxIdleConns > 0 {
		options = append(options, governpostgres.WithMaxIdleConns(cfg.MaxIdleConns))
	}
	if cfg.MaxOpenConns > 0 {
		options = append(options, governpostgres.WithMaxOpenConns(cfg.MaxOpenConns))
	}
	if cfg.MaxLifetime > 0 {
		options = append(options, governpostgres.WithConnMaxLifetime(cfg.MaxLifetime))
	}
	if cfg.MaxIdleTime > 0 {
		options = append(options, governpostgres.WithConnMaxIdleTime(cfg.MaxIdleTime))
	}

	// Use govern postgres with connection pooling
	db, cleanup, err := governpostgres.New(cfg.DSN, options...)
	if err != nil {
		return nil, nil, err
	}

	return db, cleanup, nil
}
