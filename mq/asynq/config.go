// Package asynq provides a govern-integrated wrapper around hibiken/asynq
// message queue library.
//
// # Phase Progression
//
// Phase 1: Core package structure (config, options)
// Phase 2: Client implementation
// Phase 3: Server and handler implementation
// Phase 4: Middleware
// Phase 5: Task builders
// Phase 6: Documentation
//
// # Integration Points
//
// - Redis: Uses github.com/haipham22/govern/database/redis package
// - Logging: Uses go.uber.org/zap logger from log package
// - Lifecycle: Implements graceful.Service interface for server
// - DI: Wire provider functions for dependency injection
//
// # Usage Example (Phase 2+)
//
//	client, cleanup, _ := asynq.NewClient(redisClient)
//	defer cleanup()
//
//	task := asynq.NewTask("email:send", payload)
//	info, _ := client.Enqueue(ctx, task)
package asynq

import "time"

const (
	// DefaultConcurrency is default worker concurrency.
	DefaultConcurrency = 10

	// DefaultQueue is default queue name.
	DefaultQueue = "default"

	// DefaultMaxRetry is default max retry attempts.
	DefaultMaxRetry = 25

	// DefaultTimeout is default task timeout.
	DefaultTimeout = 30 * time.Minute

	// DefaultShutdownTimeout is default graceful shutdown timeout.
	DefaultShutdownTimeout = 30 * time.Second
)

// Config holds asynq configuration.
type Config struct {
	// Concurrency number of workers to process tasks.
	Concurrency int

	// Queues mapping of queue names to priorities.
	Queues map[string]int

	// ShutdownTimeout for graceful shutdown.
	ShutdownTimeout time.Duration
}

// ClientConfig holds client configuration.
type ClientConfig struct {
	// MaxRetry attempts before giving up.
	MaxRetry int

	// Timeout for each task.
	Timeout time.Duration
}

// EnqueueConfig holds task enqueue options.
type EnqueueConfig struct {
	// Queue name for the task.
	Queue string

	// MaxRetry attempts.
	MaxRetry int

	// Timeout for task execution.
	Timeout time.Duration

	// Deadline absolute time for task.
	Deadline time.Time

	// Unique TTL for task deduplication.
	UniqueTTL time.Duration

	// ProcessIn delays task execution.
	ProcessIn time.Duration

	// ProcessAt schedules task at specific time.
	ProcessAt time.Time
}

// defaultConfig returns default config.
func defaultConfig() *Config {
	return &Config{
		Concurrency:     DefaultConcurrency,
		Queues:          map[string]int{DefaultQueue: 1},
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

// defaultClientConfig returns default client config.
func defaultClientConfig() *ClientConfig {
	return &ClientConfig{
		MaxRetry: DefaultMaxRetry,
		Timeout:  DefaultTimeout,
	}
}

// defaultEnqueueConfig returns default enqueue config.
func defaultEnqueueConfig() *EnqueueConfig {
	return &EnqueueConfig{
		Queue:    DefaultQueue,
		MaxRetry: DefaultMaxRetry,
		Timeout:  DefaultTimeout,
	}
}
