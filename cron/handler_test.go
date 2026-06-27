package cron

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// TestHandler implementation for testing
type TestHandler struct {
	setupCalled   bool
	executeCalled bool
	cleanupCalled bool
	setupError    error
	executeError  error
	cleanupError  error
}

func (h *TestHandler) Setup(session SchedulerSession) error {
	h.setupCalled = true
	return h.setupError
}

func (h *TestHandler) Execute(session SchedulerSession) error {
	h.executeCalled = true
	return h.executeError
}

func (h *TestHandler) Cleanup(session SchedulerSession) error {
	h.cleanupCalled = true
	return h.cleanupError
}

func TestJobHandlerFunc(t *testing.T) {
	t.Run("JobHandlerFunc implements JobHandler interface", func(t *testing.T) {
		// Compile-time check
		var _ JobHandler = JobHandlerFunc(nil)
	})

	t.Run("Setup returns nil (no-op)", func(t *testing.T) {
		handler := JobHandlerFunc(func(ctx context.Context) error {
			return nil
		})
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Setup(session)
		assert.NoError(t, err)
	})

	t.Run("Execute calls the underlying function", func(t *testing.T) {
		called := false
		handler := JobHandlerFunc(func(ctx context.Context) error {
			called = true
			return nil
		})
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Execute(session)
		assert.NoError(t, err)
		assert.True(t, called, "Execute should call the function")
	})

	t.Run("Execute returns function error", func(t *testing.T) {
		expectedErr := errors.New("function error")
		handler := JobHandlerFunc(func(ctx context.Context) error {
			return expectedErr
		})
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Execute(session)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Execute passes session context to function", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), contextKey("key"), "value")
		handler := JobHandlerFunc(func(ctx context.Context) error {
			assert.Equal(t, "value", ctx.Value(contextKey("key")))
			return nil
		})
		session := NewSchedulerSession(ctx, "test", "job1", time.Time{}, time.Now())

		err := handler.Execute(session)
		assert.NoError(t, err)
	})

	t.Run("Cleanup returns nil (no-op)", func(t *testing.T) {
		handler := JobHandlerFunc(func(ctx context.Context) error {
			return nil
		})
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Cleanup(session)
		assert.NoError(t, err)
	})
}

func TestCustomJobHandler(t *testing.T) {
	t.Run("full lifecycle calls all methods", func(t *testing.T) {
		handler := &TestHandler{}

		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		// Setup
		err := handler.Setup(session)
		require.NoError(t, err)
		assert.True(t, handler.setupCalled, "Setup should be called")

		// Execute
		err = handler.Execute(session)
		require.NoError(t, err)
		assert.True(t, handler.executeCalled, "Execute should be called")

		// Cleanup
		err = handler.Cleanup(session)
		require.NoError(t, err)
		assert.True(t, handler.cleanupCalled, "Cleanup should be called")
	})

	t.Run("Setup error propagates", func(t *testing.T) {
		handler := &TestHandler{
			setupError: errors.New("setup failed"),
		}
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Setup(session)
		assert.ErrorIs(t, err, handler.setupError)
	})

	t.Run("Execute error propagates", func(t *testing.T) {
		handler := &TestHandler{
			executeError: errors.New("execute failed"),
		}
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Execute(session)
		assert.ErrorIs(t, err, handler.executeError)
	})

	t.Run("Cleanup error propagates", func(t *testing.T) {
		handler := &TestHandler{
			cleanupError: errors.New("cleanup failed"),
		}
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		err := handler.Cleanup(session)
		assert.ErrorIs(t, err, handler.cleanupError)
	})

	// Implement JobHandler interface for TestHandler
	t.Run("TestHandler implements JobHandler", func(t *testing.T) {
		var _ JobHandler = (*TestHandler)(nil)
	})
}
