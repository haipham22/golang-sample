package asynq

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Middleware is an alias for asynq.Handler middleware.
// Middleware wraps a handler to add pre/post processing.
type Middleware = func(asynq.Handler) asynq.Handler

// chainMiddlewares chains multiple middleware into one.
func chainMiddlewares(handlers []Middleware) Middleware {
	return func(h asynq.Handler) asynq.Handler {
		for i := len(handlers) - 1; i >= 0; i-- {
			h = handlers[i](h)
		}
		return h
	}
}

// LoggingMiddleware logs task execution with timing.
// Logs task start, completion, and errors with duration.
func LoggingMiddleware(logger *zap.SugaredLogger) Middleware {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()
			taskType := t.Type()

			logger.Infof("Processing task: type=%s", taskType)

			err := h.ProcessTask(ctx, t)

			duration := time.Since(start)
			if err != nil {
				logger.Errorf("Task failed: type=%s duration=%s error=%v",
					taskType, duration, err)
			} else {
				logger.Infof("Task completed: type=%s duration=%s",
					taskType, duration)
			}

			return err
		})
	}
}

// RecoveryMiddleware recovers from panics in task handlers.
// Converts panics to errors to prevent server crashes.
func RecoveryMiddleware(logger *zap.SugaredLogger) Middleware {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			taskType := t.Type()

			defer func() {
				if r := recover(); r != nil {
					// Convert panic to error
					panicErr := fmt.Errorf("panic in task %s: %v", taskType, r)
					logger.Errorf("Task panic recovered: type=%s panic=%v",
						taskType, r)
					err = panicErr
				}
			}()

			return h.ProcessTask(ctx, t)
		})
	}
}

// MetricsMiddleware records task execution metrics.
// Placeholder for Prometheus metrics integration.
func MetricsMiddleware() Middleware {
	// TODO: Phase 5 - Add Prometheus metrics
	// - Tasks processed total (by type, status)
	// - Task duration histogram
	// - Tasks retry count
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return h.ProcessTask(ctx, t)
		})
	}
}

// Chain creates a middleware chain from multiple middleware.
// Middleware are applied in order (first middleware wraps the handler first).
func Chain(middleware ...Middleware) Middleware {
	return chainMiddlewares(middleware)
}

// DefaultMiddleware returns the recommended middleware stack.
// Includes logging, recovery, and metrics (placeholder).
func DefaultMiddleware(logger *zap.SugaredLogger) []Middleware {
	return []Middleware{
		RecoveryMiddleware(logger),
		LoggingMiddleware(logger),
		MetricsMiddleware(),
	}
}
