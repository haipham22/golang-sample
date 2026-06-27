package cron

import (
	"time"

	"go.uber.org/zap"
)

// Option configures a Scheduler
type Option func(*Config) error

// Config holds scheduler configuration
type Config struct {
	Logger      *zap.SugaredLogger
	Location    *time.Location
	StopTimeout time.Duration
}

// WithLogger sets the logger
func WithLogger(logger *zap.SugaredLogger) Option {
	return func(cfg *Config) error {
		cfg.Logger = logger
		return nil
	}
}

// WithLocation sets the scheduler location/timezone
func WithLocation(loc *time.Location) Option {
	return func(cfg *Config) error {
		cfg.Location = loc
		return nil
	}
}

// WithStopTimeout sets shutdown timeout
func WithStopTimeout(d time.Duration) Option {
	return func(cfg *Config) error {
		cfg.StopTimeout = d
		return nil
	}
}
