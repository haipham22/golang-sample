package asynq

import (
	"context"
	"errors"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskHandlerFunc(t *testing.T) {
	t.Run("implements TaskHandler", func(t *testing.T) {
		var _ TaskHandler = TaskHandlerFunc(nil)
	})

	t.Run("calls wrapped function", func(t *testing.T) {
		called := false
		fn := TaskHandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			called = true
			return nil
		})

		err := fn.ProcessTask(context.Background(), &asynq.Task{})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("returns error from wrapped function", func(t *testing.T) {
		expectedErr := errors.New("test error")
		fn := TaskHandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			return expectedErr
		})

		err := fn.ProcessTask(context.Background(), &asynq.Task{})
		assert.Same(t, expectedErr, err)
	})
}

func TestHandlerAdapter(t *testing.T) {
	t.Run("implements TaskHandler", func(t *testing.T) {
		var _ = NewHandlerAdapter(nil)
	})

	t.Run("delegates to asynq.Handler", func(t *testing.T) {
		expectedErr := errors.New("test error")
		asynqHandler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return expectedErr
		})

		adapter := NewHandlerAdapter(asynqHandler)
		err := adapter.ProcessTask(context.Background(), &asynq.Task{})
		assert.Same(t, expectedErr, err)
	})
}

func TestBaseHandler(t *testing.T) {
	t.Run("new base handler", func(t *testing.T) {
		h := NewBaseHandler()
		assert.NotNil(t, h)
	})

	t.Run("parse valid payload", func(t *testing.T) {
		h := NewBaseHandler()

		payload := map[string]interface{}{"key": "value"}
		data, err := sonic.Marshal(payload)
		require.NoError(t, err)

		task := asynq.NewTask("test", data)

		var result map[string]interface{}
		err = h.ParsePayload(task, &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("parse invalid payload", func(t *testing.T) {
		h := NewBaseHandler()

		task := asynq.NewTask("test", []byte("invalid json"))

		var result map[string]interface{}
		err := h.ParsePayload(task, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse payload")
	})

	t.Run("parse into struct", func(t *testing.T) {
		h := NewBaseHandler()

		type TestPayload struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		payload := TestPayload{Name: "test", Count: 42}
		data, err := sonic.Marshal(payload)
		require.NoError(t, err)

		task := asynq.NewTask("test", data)

		var result TestPayload
		err = h.ParsePayload(task, &result)
		assert.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Count)
	})
}

func TestParsePayload(t *testing.T) {
	t.Run("valid payload", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		data, err := sonic.Marshal(payload)
		require.NoError(t, err)

		task := asynq.NewTask("test", data)

		var result map[string]string
		err = ParsePayload(task, &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("invalid payload", func(t *testing.T) {
		task := asynq.NewTask("test", []byte("invalid json"))

		var result map[string]string
		err := ParsePayload(task, &result)
		assert.Error(t, err)
	})
}

func TestTaskError(t *testing.T) {
	t.Run("new task error", func(t *testing.T) {
		task := asynq.NewTask("test:type", nil)
		originalErr := errors.New("original error")

		taskErr := NewTaskError(task, originalErr)

		assert.Error(t, taskErr)
		assert.Contains(t, taskErr.Error(), "test:type")
		assert.Contains(t, taskErr.Error(), "original error")
	})

	t.Run("nil error returns nil", func(t *testing.T) {
		task := asynq.NewTask("test:type", nil)
		taskErr := NewTaskError(task, nil)
		assert.NoError(t, taskErr)
	})

	t.Run("unwrap returns original error", func(t *testing.T) {
		task := asynq.NewTask("test:type", nil)
		originalErr := errors.New("original error")

		taskErr := NewTaskError(task, originalErr)

		unwrapped := errors.Unwrap(taskErr)
		assert.Same(t, originalErr, unwrapped)
	})
}
