package asynq

import (
	"context"
	"sync"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewTaskMux(t *testing.T) {
	t.Run("creates mux without options", func(t *testing.T) {
		mux := NewTaskMux()
		assert.NotNil(t, mux)
	})

	t.Run("creates mux with logger", func(t *testing.T) {
		logger := zap.NewNop().Sugar()
		mux := NewTaskMux(WithMuxLogger(logger))
		assert.NotNil(t, mux)
	})
}

func TestTaskMux_Handle(t *testing.T) {
	t.Run("register handler", func(t *testing.T) {
		mux := NewTaskMux()
		handler := &muxTestHandler{}

		mux.Handle("test:task", handler)

		assert.True(t, mux.HasHandler("test:task"))
		assert.Equal(t, []string{"test:task"}, mux.HandlerTypes())
	})

	t.Run("panic on nil handler", func(t *testing.T) {
		mux := NewTaskMux()

		assert.Panics(t, func() {
			mux.Handle("test:task", nil)
		})
	})

	t.Run("panic on duplicate handler", func(t *testing.T) {
		mux := NewTaskMux()
		handler := &muxTestHandler{}

		mux.Handle("test:task", handler)

		assert.Panics(t, func() {
			mux.Handle("test:task", handler)
		})
	})
}

func TestTaskMux_HandleFunc(t *testing.T) {
	t.Run("register function handler", func(t *testing.T) {
		mux := NewTaskMux()

		mux.HandleFunc("test:task", func(ctx context.Context, task *asynq.Task) error {
			return nil
		})

		assert.True(t, mux.HasHandler("test:task"))
	})

	t.Run("panic on nil function", func(t *testing.T) {
		mux := NewTaskMux()

		assert.Panics(t, func() {
			mux.HandleFunc("test:task", nil)
		})
	})

	t.Run("calls function handler", func(t *testing.T) {
		mux := NewTaskMux()
		called := false

		mux.HandleFunc("test:task", func(ctx context.Context, task *asynq.Task) error {
			called = true
			return nil
		})

		task := asynq.NewTask("test:task", nil)
		err := mux.HandleTask(context.Background(), task)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestTaskMux_HandleTask(t *testing.T) {
	t.Run("routes to registered handler", func(t *testing.T) {
		mux := NewTaskMux()
		handler := &muxTestHandler{result: nil}

		mux.Handle("test:task", handler)

		task := asynq.NewTask("test:task", nil)
		err := mux.HandleTask(context.Background(), task)

		assert.NoError(t, err)
		assert.True(t, handler.called)
	})

	t.Run("returns error for nil task", func(t *testing.T) {
		mux := NewTaskMux()
		mux.Handle("test:task", &muxTestHandler{})

		err := mux.HandleTask(context.Background(), nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task is nil")
	})

	t.Run("returns error for unregistered handler", func(t *testing.T) {
		mux := NewTaskMux()

		task := asynq.NewTask("unknown:task", nil)
		err := mux.HandleTask(context.Background(), task)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no handler registered")
	})

	t.Run("returns handler error", func(t *testing.T) {
		mux := NewTaskMux()
		expectedErr := assert.AnError
		handler := &muxTestHandler{result: expectedErr}

		mux.Handle("test:task", handler)

		task := asynq.NewTask("test:task", nil)
		err := mux.HandleTask(context.Background(), task)

		assert.Same(t, expectedErr, err)
	})
}

func TestTaskMux_ServeAsynqHandler(t *testing.T) {
	t.Run("implements asynq.Handler", func(t *testing.T) {
		var _ asynq.Handler = NewTaskMux()
	})

	t.Run("serves as asynq handler", func(t *testing.T) {
		mux := NewTaskMux()
		handler := &muxTestHandler{result: nil}

		mux.Handle("test:task", handler)

		task := asynq.NewTask("test:task", nil)
		err := mux.ServeAsynqHandler(context.Background(), task)

		assert.NoError(t, err)
		assert.True(t, handler.called)
	})
}

func TestTaskMux_HandlerTypes(t *testing.T) {
	t.Run("returns empty list when no handlers", func(t *testing.T) {
		mux := NewTaskMux()
		types := mux.HandlerTypes()
		assert.Empty(t, types)
	})

	t.Run("returns all registered types", func(t *testing.T) {
		mux := NewTaskMux()
		mux.Handle("task1", &muxTestHandler{})
		mux.Handle("task2", &muxTestHandler{})
		mux.Handle("task3", &muxTestHandler{})

		types := mux.HandlerTypes()
		assert.Len(t, types, 3)
		assert.ElementsMatch(t, []string{"task1", "task2", "task3"}, types)
	})
}

func TestTaskMux_HasHandler(t *testing.T) {
	t.Run("returns false for unregistered type", func(t *testing.T) {
		mux := NewTaskMux()
		assert.False(t, mux.HasHandler("unknown"))
	})

	t.Run("returns true for registered type", func(t *testing.T) {
		mux := NewTaskMux()
		mux.Handle("test", &muxTestHandler{})
		assert.True(t, mux.HasHandler("test"))
	})
}

func TestTaskMux_Unregister(t *testing.T) {
	t.Run("removes registered handler", func(t *testing.T) {
		mux := NewTaskMux()
		mux.Handle("test", &muxTestHandler{})

		assert.True(t, mux.HasHandler("test"))
		assert.True(t, mux.Unregister("test"))
		assert.False(t, mux.HasHandler("test"))
	})

	t.Run("returns false for unregistered type", func(t *testing.T) {
		mux := NewTaskMux()
		assert.False(t, mux.Unregister("unknown"))
	})
}

func TestTaskMux_Concurrent(t *testing.T) {
	t.Run("concurrent handle calls", func(t *testing.T) {
		mux := NewTaskMux()
		handler := &muxTestHandler{result: nil}
		mux.Handle("test", handler)

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				task := asynq.NewTask("test", nil)
				_ = mux.HandleTask(context.Background(), task)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		// Handler should have been called (use atomic read)
		assert.True(t, handler.WasCalled())
	})
}

// muxTestHandler is a test implementation of TaskHandler.
type muxTestHandler struct {
	mu     sync.Mutex
	called bool
	result error
}

func (m *muxTestHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
	return m.result
}

func (m *muxTestHandler) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}
