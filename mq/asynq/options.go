package asynq

import (
	"time"

	"go.uber.org/zap"
)

// Option configures asynq server.
type Option func(*Config)

// WithConcurrency sets worker concurrency.
func WithConcurrency(n int) Option {
	return func(cfg *Config) {
		if n > 0 {
			cfg.Concurrency = n
		}
	}
}

// WithQueues sets queue priorities.
// Example: WithQueues(map[string]int{"critical": 5, "default": 1})
func WithQueues(queues map[string]int) Option {
	return func(cfg *Config) {
		if len(queues) > 0 {
			cfg.Queues = queues
		}
	}
}

// WithShutdownTimeout sets graceful shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(cfg *Config) {
		if d > 0 {
			cfg.ShutdownTimeout = d
		}
	}
}

// ClientOption configures asynq client.
type ClientOption func(*ClientConfig)

// WithClientMaxRetry sets max retry attempts for client.
func WithClientMaxRetry(n int) ClientOption {
	return func(cfg *ClientConfig) {
		if n > 0 {
			cfg.MaxRetry = n
		}
	}
}

// WithClientTimeout sets default timeout for tasks.
func WithClientTimeout(d time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		if d > 0 {
			cfg.Timeout = d
		}
	}
}

// WithLogger sets Zap logger (used for both client and server).
// TODO: Phase 2 - Store logger in Config struct and use in Client/Server initialization.
func WithLogger(logger *zap.SugaredLogger) Option {
	return func(cfg *Config) {
		// Logger is stored at package level for middleware
		// Set via initLogger when server/client is created
		_ = logger
	}
}

// EnqueueOption configures task enqueue behavior.
type EnqueueOption func(*EnqueueConfig)

// WithQueue sets target queue name.
func WithQueue(name string) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		if name != "" {
			cfg.Queue = name
		}
	}
}

// WithMaxRetry sets max retry attempts for task.
func WithMaxRetry(n int) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		if n >= 0 {
			cfg.MaxRetry = n
		}
	}
}

// WithEnqueueTimeout sets task timeout.
func WithEnqueueTimeout(d time.Duration) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		if d > 0 {
			cfg.Timeout = d
		}
	}
}

// WithDeadline sets absolute deadline for task.
func WithDeadline(t time.Time) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		if !t.IsZero() {
			cfg.Deadline = t
		}
	}
}

// WithUnique enables task deduplication with TTL.
func WithUnique(ttl time.Duration) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		cfg.UniqueTTL = ttl
	}
}

// WithProcessIn delays task execution.
func WithProcessIn(d time.Duration) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		cfg.ProcessIn = d
	}
}

// WithProcessAt schedules task at specific time.
func WithProcessAt(t time.Time) EnqueueOption {
	return func(cfg *EnqueueConfig) {
		if !t.IsZero() {
			cfg.ProcessAt = t
		}
	}
}

// MuxOption configures TaskMux.
type MuxOption func(*muxConfig)

type muxConfig struct {
	logger *zap.SugaredLogger
}

// WithMuxLogger sets logger for TaskMux.
func WithMuxLogger(logger *zap.SugaredLogger) MuxOption {
	return func(cfg *muxConfig) {
		cfg.logger = logger
	}
}

// MiddlewareOption adds middleware to server.
type MiddlewareOption interface {
	Option // Embedded Option to allow WithMiddleware([]MiddlewareFunc)
}

// MiddlewareFunc represents asynq middleware function.
// Type alias for compatibility with hibiken/asynq middleware.
type MiddlewareFunc = interface{}
