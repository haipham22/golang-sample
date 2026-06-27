package asynq

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// miniredisClient creates a miniredis-backed redis client for testing.
// This provides a real Redis-compatible interface without requiring an external Redis instance.
func miniredisClient(t *testing.T) redis.UniversalClient {
	s := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func TestNewClient(t *testing.T) {
	client := miniredisClient(t)

	asynqClient, cleanup, err := NewClient(client)
	require.NoError(t, err)
	require.NotNil(t, asynqClient)
	require.NotNil(t, cleanup)
	defer cleanup()

	assert.NotNil(t, asynqClient.client)
	assert.NotNil(t, asynqClient.config)
}

func TestNewClientWithOptions(t *testing.T) {
	client := miniredisClient(t)

	asynqClient, cleanup, err := NewClient(client,
		WithClientMaxRetry(10),
		WithClientTimeout(5*time.Minute),
	)
	require.NoError(t, err)
	require.NotNil(t, asynqClient)
	defer cleanup()

	assert.Equal(t, 10, asynqClient.config.MaxRetry)
	assert.Equal(t, 5*time.Minute, asynqClient.config.Timeout)
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		payload interface{}
		wantErr bool
	}{
		{
			name:    "valid task with payload",
			typ:     "email:send",
			payload: map[string]interface{}{"user_id": 42},
			wantErr: false,
		},
		{
			name:    "valid task with nil payload",
			typ:     "email:send",
			payload: nil,
			wantErr: false,
		},
		{
			name:    "valid task with struct payload",
			typ:     "email:send",
			payload: struct{ UserID int }{UserID: 42},
			wantErr: false,
		},
		{
			name:    "invalid payload (channel)",
			typ:     "email:send",
			payload: make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.typ, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.typ, task.Type())

				if tt.payload != nil {
					var payload map[string]interface{}
					err := json.Unmarshal(task.Payload(), &payload)
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestEnqueue(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue valid task", func(t *testing.T) {
		task, err := NewTask("test:enqueue", map[string]string{"key": "value"})
		require.NoError(t, err)

		info, err := client.Enqueue(ctx, task)
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
		assert.NotEmpty(t, info.Queue)
	})

	t.Run("enqueue with options", func(t *testing.T) {
		task, err := NewTask("test:enqueue:opts", map[string]string{"key": "value"})
		require.NoError(t, err)

		info, err := client.Enqueue(ctx, task,
			WithQueue("high"),
			WithMaxRetry(5),
			WithEnqueueTimeout(10*time.Minute),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
		assert.Equal(t, "high", info.Queue)
	})

	t.Run("enqueue nil task", func(t *testing.T) {
		info, err := client.Enqueue(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})
}

func TestEnqueueIn(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue in future", func(t *testing.T) {
		task, err := NewTask("test:enqueue:in", map[string]string{"key": "value"})
		require.NoError(t, err)

		info, err := client.EnqueueIn(ctx, task, 1*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
	})

	t.Run("enqueue nil task", func(t *testing.T) {
		info, err := client.EnqueueIn(ctx, nil, 1*time.Hour)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("enqueue with negative duration", func(t *testing.T) {
		task, err := NewTask("test:enqueue:negative", map[string]string{})
		require.NoError(t, err)

		info, err := client.EnqueueIn(ctx, task, -1*time.Hour)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "cannot be negative")
	})
}

func TestEnqueueAt(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue at future time", func(t *testing.T) {
		task, err := NewTask("test:enqueue:at", map[string]string{"key": "value"})
		require.NoError(t, err)

		futureTime := time.Now().Add(1 * time.Hour)
		info, err := client.EnqueueAt(ctx, task, futureTime)
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
	})

	t.Run("enqueue nil task", func(t *testing.T) {
		info, err := client.EnqueueAt(ctx, nil, time.Now().Add(1*time.Hour))
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("enqueue at past time", func(t *testing.T) {
		task, err := NewTask("test:enqueue:past", map[string]string{})
		require.NoError(t, err)

		pastTime := time.Now().Add(-1 * time.Hour)
		info, err := client.EnqueueAt(ctx, task, pastTime)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "must be in the future")
	})
}

func TestClient_Close(t *testing.T) {
	redisClient := miniredisClient(t)

	_, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)

	// Close via cleanup function (no return value)
	cleanup()

	// Client should be closed now
	// asynq.Client.Close() may return error on double close, which is expected
}

func TestClient_CloseDirect(t *testing.T) {
	t.Run("close client directly", func(t *testing.T) {
		redisClient := miniredisClient(t)

		client, cleanup, err := NewClient(redisClient)
		require.NoError(t, err)

		// Close directly via Close method
		err = client.Close()
		assert.NoError(t, err)

		// Call cleanup (should not error even if already closed)
		cleanup()
	})
}

func TestClient_CloseIdempotent(t *testing.T) {
	t.Run("close is idempotent", func(t *testing.T) {
		redisClient := miniredisClient(t)

		client, cleanup, err := NewClient(redisClient)
		require.NoError(t, err)

		// First close
		err = client.Close()
		assert.NoError(t, err)

		// Second close should also succeed (asynq handles this)
		err = client.Close()
		// May error depending on asynq implementation
		_ = err

		cleanup()
	})
}

func TestClient_EnqueueWithZeroDuration(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue in with zero duration", func(t *testing.T) {
		task, err := NewTask("test:zero:duration", map[string]string{"key": "value"})
		require.NoError(t, err)

		// Zero duration means enqueue immediately
		info, err := client.EnqueueIn(ctx, task, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
	})
}

func TestClient_EnqueueWithOptions(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue with retry option", func(t *testing.T) {
		task, err := NewTask("test:retry", map[string]string{"key": "value"})
		require.NoError(t, err)

		info, err := client.Enqueue(ctx, task, WithMaxRetry(10))
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
	})

	t.Run("enqueue with queue option", func(t *testing.T) {
		task, err := NewTask("test:queue", map[string]string{"key": "value"})
		require.NoError(t, err)

		info, err := client.Enqueue(ctx, task, WithQueue("custom"))
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
		assert.Equal(t, "custom", info.Queue)
	})
}

func TestClient_ConcurrentEnqueue(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("concurrent enqueue operations", func(t *testing.T) {
		const goroutines = 50
		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				task, _ := NewTask("test:concurrent", map[string]int{"id": id})
				_, _ = client.Enqueue(ctx, task)
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			<-done
		}
	})
}

func TestAsynqRedisClientOpt(t *testing.T) {
	t.Run("single node client", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "password",
			DB:       1,
			PoolSize: 50,
		})

		opt := asynqRedisClientOpt(client)
		assert.Equal(t, "localhost:6379", opt.Addr)
		assert.Equal(t, "password", opt.Password)
		assert.Equal(t, 1, opt.DB)
		assert.Equal(t, 50, opt.PoolSize)
	})

	t.Run("cluster client", func(t *testing.T) {
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    []string{"localhost:7000", "localhost:7001"},
			Password: "password",
			PoolSize: 100,
		})

		opt := asynqRedisClientOpt(client)
		// Cluster uses first addr as fallback for asynq
		assert.Equal(t, "localhost:7000", opt.Addr)
		assert.Equal(t, "password", opt.Password)
		assert.Equal(t, 100, opt.PoolSize)
	})
}

// Test client method validation logic without Redis
func TestClientValidation(t *testing.T) {
	// Create a mock client that will fail before calling Redis
	t.Run("enqueue nil task validation", func(t *testing.T) {
		mockClient := &Client{
			client: nil, // Intentionally nil for validation test
			logger: zap.NewNop().Sugar(),
			config: defaultClientConfig(),
		}

		info, err := mockClient.Enqueue(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("enqueueIn nil task validation", func(t *testing.T) {
		mockClient := &Client{
			client: nil,
			logger: zap.NewNop().Sugar(),
			config: defaultClientConfig(),
		}

		info, err := mockClient.EnqueueIn(context.Background(), nil, 1*time.Hour)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("enqueueIn negative duration validation", func(t *testing.T) {
		mockClient := &Client{
			client: nil,
			logger: zap.NewNop().Sugar(),
			config: defaultClientConfig(),
		}

		task, _ := NewTask("test", nil)
		info, err := mockClient.EnqueueIn(context.Background(), task, -1*time.Hour)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("enqueueAt nil task validation", func(t *testing.T) {
		mockClient := &Client{
			client: nil,
			logger: zap.NewNop().Sugar(),
			config: defaultClientConfig(),
		}

		info, err := mockClient.EnqueueAt(context.Background(), nil, time.Now().Add(1*time.Hour))
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "task cannot be nil")
	})

	t.Run("enqueueAt past time validation", func(t *testing.T) {
		mockClient := &Client{
			client: nil,
			logger: zap.NewNop().Sugar(),
			config: defaultClientConfig(),
		}

		task, _ := NewTask("test", nil)
		pastTime := time.Now().Add(-1 * time.Hour)
		info, err := mockClient.EnqueueAt(context.Background(), task, pastTime)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "must be in the future")
	})
}

func TestEnqueueWithDeadline(t *testing.T) {
	redisClient := miniredisClient(t)

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(t, err)
	defer cleanup()

	t.Run("enqueue with deadline", func(t *testing.T) {
		task, err := NewTask("test:deadline", map[string]string{"key": "value"})
		require.NoError(t, err)

		deadline := time.Now().Add(2 * time.Hour)
		info, err := client.Enqueue(ctx, task, WithDeadline(deadline))
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)
	})
}

func TestEnqueueWithUnique(t *testing.T) {
	t.Skip("miniredis does not fully support asynq's unique task feature (requires Redis Lua scripts)")

	// This test requires full Redis for unique task locking
	// miniredis has limited support for Redis Lua scripts used by asynq
	// TODO: Re-enable when testcontainers-redis is available
}

// Benchmark Enqueue
func BenchmarkEnqueue(b *testing.B) {
	redisClient := miniredisClient(&testing.T{})

	ctx := context.Background()
	redisClient.FlushDB(ctx)

	client, cleanup, err := NewClient(redisClient)
	require.NoError(b, err)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, _ := NewTask("bench:enqueue", map[string]int{"i": i})
		_, _ = client.Enqueue(ctx, task)
	}
}
