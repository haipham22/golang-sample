package asynq

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs task processing", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := LoggingMiddleware(logger)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})

	t.Run("logs task errors", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := LoggingMiddleware(logger)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return errors.New("task error")
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
	})

	t.Run("handles nil logger", func(t *testing.T) {
		middleware := LoggingMiddleware(nil)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := RecoveryMiddleware(logger)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			panic("intentional panic")
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic")
		assert.Contains(t, err.Error(), "intentional panic")
	})

	t.Run("passes through normal errors", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := RecoveryMiddleware(logger)

		expectedErr := errors.New("task error")
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return expectedErr
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Same(t, expectedErr, err)
	})

	t.Run("passes through success", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := RecoveryMiddleware(logger)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})

	t.Run("handles nil logger", func(t *testing.T) {
		middleware := RecoveryMiddleware(nil)

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			panic("test")
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
	})
}

func TestMetricsMiddleware(t *testing.T) {
	t.Run("placeholder middleware", func(t *testing.T) {
		middleware := MetricsMiddleware()

		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		wrapped := middleware(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})
}

func TestChain(t *testing.T) {
	t.Run("chains multiple middleware", func(t *testing.T) {
		callOrder := []string{}
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			callOrder = append(callOrder, "handler")
			return nil
		})

		middleware1 := func(h asynq.Handler) asynq.Handler {
			return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
				callOrder = append(callOrder, "middleware1-before")
				err := h.ProcessTask(ctx, t)
				callOrder = append(callOrder, "middleware1-after")
				return err
			})
		}

		middleware2 := func(h asynq.Handler) asynq.Handler {
			return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
				callOrder = append(callOrder, "middleware2-before")
				err := h.ProcessTask(ctx, t)
				callOrder = append(callOrder, "middleware2-after")
				return err
			})
		}

		chain := Chain(middleware1, middleware2)
		wrapped := chain(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
		assert.Equal(t, []string{
			"middleware1-before",
			"middleware2-before",
			"handler",
			"middleware2-after",
			"middleware1-after",
		}, callOrder)
	})

	t.Run("empty chain returns handler", func(t *testing.T) {
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		chain := Chain()
		wrapped := chain(handler)

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})
}

func TestDefaultMiddleware(t *testing.T) {
	t.Run("returns default stack", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := DefaultMiddleware(logger)

		assert.Len(t, middleware, 3)
	})
}

func TestMiddlewareIntegration(t *testing.T) {
	t.Run("logging and recovery middleware", func(t *testing.T) {
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return nil
		})

		middlewares := []Middleware{
			RecoveryMiddleware(nil),
			LoggingMiddleware(nil),
		}

		// Apply middleware
		var wrapped asynq.Handler = handler
		for i := len(middlewares) - 1; i >= 0; i-- {
			wrapped = middlewares[i](wrapped)
		}

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.NoError(t, err)
	})

	t.Run("panic is recovered and logged", func(t *testing.T) {
		handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			panic("test panic")
		})

		middlewares := []Middleware{
			RecoveryMiddleware(nil),
			LoggingMiddleware(nil),
		}

		var wrapped asynq.Handler = handler
		for i := len(middlewares) - 1; i >= 0; i-- {
			wrapped = middlewares[i](wrapped)
		}

		task := asynq.NewTask("test:task", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic")
	})
}

// mockHandler is a test handler that can be configured to panic or error.
type mockHandler struct {
	shouldPanic bool
	shouldError bool
}

func (m *mockHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if m.shouldPanic {
		panic("mock panic")
	}
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

func TestMiddlewareWithMockHandler(t *testing.T) {
	t.Run("recovery catches mock panic", func(t *testing.T) {
		logger := zap.NewExample().Sugar()
		middleware := RecoveryMiddleware(logger)

		handler := &mockHandler{shouldPanic: true}
		wrapped := middleware(handler)

		task := asynq.NewTask("test", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic")
	})

	t.Run("middleware passes through errors", func(t *testing.T) {
		middleware := RecoveryMiddleware(nil)

		handler := &mockHandler{shouldError: true}
		wrapped := middleware(handler)

		task := asynq.NewTask("test", nil)
		err := wrapped.ProcessTask(context.Background(), task)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}
