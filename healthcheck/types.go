package healthcheck

import (
	"context"
	"time"
)

// Status represents the health check result status.
type Status string

const (
	StatusPassing Status = "pass"
	StatusFailing Status = "fail"
	StatusWarning Status = "warn"
)

// Check is a health check function.
type Check func(ctx context.Context) error

// Result represents the result of a single health check.
type Result struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration_ms,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Response is the overall health check response.
type Response struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Result `json:"checks,omitempty"`
	Duration  time.Duration     `json:"duration_ms,omitempty"`
}

// Option configures a Check registration.
type Option func(*checkConfig)

type checkConfig struct {
	timeout      time.Duration
	disablePanic bool
}

// WithTimeout sets a timeout for the health check.
func WithTimeout(d time.Duration) Option {
	return func(c *checkConfig) {
		c.timeout = d
	}
}

// DisablePanic prevents the check from panicking.
func DisablePanic() Option {
	return func(c *checkConfig) {
		c.disablePanic = true
	}
}
