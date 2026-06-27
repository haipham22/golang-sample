package asynq

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	t.Run("creates server with valid args", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		mux.HandleFunc("test", func(ctx context.Context, task *asynq.Task) error {
			return nil
		})

		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		require.NotNil(t, server)
		require.NotNil(t, cleanup)
		defer cleanup()

		assert.NotNil(t, server.server)
		assert.NotNil(t, server.mux)
		assert.NotNil(t, server.config)
	})

	t.Run("panics on nil mux", func(t *testing.T) {
		redisClient := miniredisClient(t)

		_, _, err := NewServer(redisClient, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mux cannot be nil")
	})

	t.Run("applies options", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()

		server, cleanup, err := NewServer(redisClient, mux,
			WithConcurrency(20),
			WithQueues(map[string]int{"critical": 5}),
		)
		require.NoError(t, err)
		defer cleanup()

		assert.Equal(t, 20, server.config.Concurrency)
		assert.Equal(t, 5, server.config.Queues["critical"])
	})
}

func TestServer_Start(t *testing.T) {
	t.Skip("Skipping server startup tests in miniredis environment - asynq server requires actual Redis for proper integration testing")

	t.Run("starts server successfully", func(t *testing.T) {
		// Skipped - requires actual Redis instance
	})

	t.Run("cannot start twice", func(t *testing.T) {
		// Skipped - requires actual Redis instance
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Run("shuts down server gracefully", func(t *testing.T) {
		t.Skip("Skipping server shutdown test - requires actual Redis instance")
	})

	t.Run("cannot shutdown twice", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		// First shutdown
		err = server.Shutdown(context.Background())
		assert.NoError(t, err)

		// Second shutdown should fail
		err = server.Shutdown(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already shutdown")
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("close is alias for shutdown", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		// Close should work without starting
		err = server.Close()
		assert.NoError(t, err)
		assert.True(t, server.IsShutdown())
	})
}

func TestServer_SetLogger(t *testing.T) {
	t.Run("sets logger", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		logger := zap.NewNop().Sugar()
		server.SetLogger(logger)

		// Logger should be set (no direct way to verify, but no panic is success)
	})
}

func TestServer_IsStarted(t *testing.T) {
	t.Run("returns false before start", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		assert.False(t, server.IsStarted())
	})
}

func TestServer_IsShutdown(t *testing.T) {
	t.Run("returns false before shutdown", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		assert.False(t, server.IsShutdown())
	})
}

func TestServer_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent IsStarted and IsShutdown calls", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		// Test concurrent access to state-checking methods
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_ = server.IsStarted()
				_ = server.IsShutdown()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestServer_SetLoggerConcurrent(t *testing.T) {
	t.Run("concurrent SetLogger calls", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		logger := zap.NewNop().Sugar()

		// Test concurrent SetLogger calls
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				server.SetLogger(logger)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestServer_HandlerAdapter(t *testing.T) {
	t.Run("handlerAdapter returns mux", func(t *testing.T) {
		redisClient := miniredisClient(t)

		mux := NewTaskMux()
		mux.HandleFunc("test", func(ctx context.Context, task *asynq.Task) error {
			return nil
		})

		server, cleanup, err := NewServer(redisClient, mux)
		require.NoError(t, err)
		defer cleanup()

		// handlerAdapter is a private method, but we can verify mux is set
		assert.NotNil(t, server.mux)
	})
}

func TestServer_ShutdownTimeout(t *testing.T) {
	t.Skip("Skipping server shutdown timeout test - requires actual Redis instance")
}

func TestServer_OptionsValidation(t *testing.T) {
	t.Run("WithConcurrency positive value sets", func(t *testing.T) {
		cfg := &Config{}
		WithConcurrency(10)(cfg)
		assert.Equal(t, 10, cfg.Concurrency)
	})

	t.Run("WithConcurrency zero value ignored", func(t *testing.T) {
		cfg := &Config{}
		cfg.Concurrency = 5 // Set initial value
		WithConcurrency(0)(cfg)
		// Zero value should be ignored, original remains
		assert.Equal(t, 5, cfg.Concurrency)
	})

	t.Run("WithConcurrency negative value ignored", func(t *testing.T) {
		cfg := &Config{}
		cfg.Concurrency = 5 // Set initial value
		WithConcurrency(-5)(cfg)
		// Negative value should be ignored
		assert.Equal(t, 5, cfg.Concurrency)
	})

	t.Run("WithQueues empty map ignored", func(t *testing.T) {
		cfg := &Config{}
		WithQueues(map[string]int{})(cfg)
		// Empty map should be ignored
		assert.Nil(t, cfg.Queues)
	})

	t.Run("WithQueues non-empty map sets", func(t *testing.T) {
		cfg := &Config{}
		queues := map[string]int{"critical": 5}
		WithQueues(queues)(cfg)
		assert.Equal(t, queues, cfg.Queues)
	})

	t.Run("WithShutdownTimeout positive value sets", func(t *testing.T) {
		cfg := &Config{}
		WithShutdownTimeout(5 * time.Second)(cfg)
		assert.Equal(t, 5*time.Second, cfg.ShutdownTimeout)
	})

	t.Run("WithShutdownTimeout zero value ignored", func(t *testing.T) {
		cfg := &Config{}
		cfg.ShutdownTimeout = 10 * time.Second // Set initial value
		WithShutdownTimeout(0)(cfg)
		// Zero value should be ignored
		assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	})

	t.Run("WithShutdownTimeout negative value ignored", func(t *testing.T) {
		cfg := &Config{}
		cfg.ShutdownTimeout = 10 * time.Second // Set initial value
		WithShutdownTimeout(-5 * time.Second)(cfg)
		// Negative value should be ignored
		assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	})
}

func TestClient_OptionsValidation(t *testing.T) {
	t.Run("WithClientMaxRetry positive value sets", func(t *testing.T) {
		cfg := &ClientConfig{}
		WithClientMaxRetry(10)(cfg)
		assert.Equal(t, 10, cfg.MaxRetry)
	})

	t.Run("WithClientMaxRetry zero value ignored", func(t *testing.T) {
		cfg := &ClientConfig{}
		cfg.MaxRetry = 5 // Set initial value
		WithClientMaxRetry(0)(cfg)
		assert.Equal(t, 5, cfg.MaxRetry)
	})

	t.Run("WithClientMaxRetry negative value ignored", func(t *testing.T) {
		cfg := &ClientConfig{}
		cfg.MaxRetry = 5 // Set initial value
		WithClientMaxRetry(-5)(cfg)
		assert.Equal(t, 5, cfg.MaxRetry)
	})

	t.Run("WithClientTimeout positive value sets", func(t *testing.T) {
		cfg := &ClientConfig{}
		WithClientTimeout(10 * time.Second)(cfg)
		assert.Equal(t, 10*time.Second, cfg.Timeout)
	})

	t.Run("WithClientTimeout zero value ignored", func(t *testing.T) {
		cfg := &ClientConfig{}
		cfg.Timeout = 5 * time.Second // Set initial value
		WithClientTimeout(0)(cfg)
		assert.Equal(t, 5*time.Second, cfg.Timeout)
	})

	t.Run("WithClientTimeout negative value ignored", func(t *testing.T) {
		cfg := &ClientConfig{}
		cfg.Timeout = 5 * time.Second // Set initial value
		WithClientTimeout(-5 * time.Second)(cfg)
		assert.Equal(t, 5*time.Second, cfg.Timeout)
	})
}

func TestEnqueue_OptionsValidation(t *testing.T) {
	t.Run("WithQueue non-empty string sets", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithQueue("critical")(cfg)
		assert.Equal(t, "critical", cfg.Queue)
	})

	t.Run("WithQueue empty string ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.Queue = "default" // Set initial value
		WithQueue("")(cfg)
		assert.Equal(t, "default", cfg.Queue)
	})

	t.Run("WithMaxRetry negative value ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.MaxRetry = 5 // Set initial value
		WithMaxRetry(-1)(cfg)
		// Negative value should be ignored, original remains
		assert.Equal(t, 5, cfg.MaxRetry)
	})

	t.Run("WithMaxRetry zero value allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithMaxRetry(0)(cfg)
		assert.Equal(t, 0, cfg.MaxRetry)
	})

	t.Run("WithEnqueueTimeout positive value sets", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithEnqueueTimeout(10 * time.Second)(cfg)
		assert.Equal(t, 10*time.Second, cfg.Timeout)
	})

	t.Run("WithEnqueueTimeout zero value ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.Timeout = 5 * time.Second // Set initial value
		WithEnqueueTimeout(0)(cfg)
		assert.Equal(t, 5*time.Second, cfg.Timeout)
	})

	t.Run("WithEnqueueTimeout negative value ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.Timeout = 5 * time.Second // Set initial value
		WithEnqueueTimeout(-5 * time.Second)(cfg)
		assert.Equal(t, 5*time.Second, cfg.Timeout)
	})

	t.Run("WithDeadline non-zero time sets", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		future := time.Now().Add(1 * time.Hour)
		WithDeadline(future)(cfg)
		assert.Equal(t, future, cfg.Deadline)
	})

	t.Run("WithDeadline zero time ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.Deadline = time.Now().Add(1 * time.Hour) // Set initial value
		WithDeadline(time.Time{})(cfg)
		assert.False(t, cfg.Deadline.IsZero())
	})

	t.Run("WithUnique zero TTL allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithUnique(0)(cfg)
		assert.Equal(t, time.Duration(0), cfg.UniqueTTL)
	})

	t.Run("WithUnique negative TTL allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithUnique(-5 * time.Second)(cfg)
		assert.Equal(t, -5*time.Second, cfg.UniqueTTL)
	})

	t.Run("WithUnique positive TTL allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithUnique(1 * time.Hour)(cfg)
		assert.Equal(t, 1*time.Hour, cfg.UniqueTTL)
	})

	t.Run("WithProcessIn zero allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithProcessIn(0)(cfg)
		assert.Equal(t, time.Duration(0), cfg.ProcessIn)
	})

	t.Run("WithProcessIn negative allowed", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		WithProcessIn(-5 * time.Second)(cfg)
		assert.Equal(t, -5*time.Second, cfg.ProcessIn)
	})

	t.Run("WithProcessAt non-zero time sets", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		future := time.Now().Add(1 * time.Hour)
		WithProcessAt(future)(cfg)
		assert.Equal(t, future, cfg.ProcessAt)
	})

	t.Run("WithProcessAt zero time ignored", func(t *testing.T) {
		cfg := &EnqueueConfig{}
		cfg.ProcessAt = time.Now().Add(1 * time.Hour) // Set initial value
		WithProcessAt(time.Time{})(cfg)
		assert.False(t, cfg.ProcessAt.IsZero())
	})
}
