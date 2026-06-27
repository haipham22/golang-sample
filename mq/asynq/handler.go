package asynq

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
)

// TaskHandler wraps asynq.Handler with govern patterns.
// Handlers process tasks asynchronously from the queue.
type TaskHandler interface {
	// ProcessTask handles the task execution.
	// Any error returned will trigger retry based on asynq configuration.
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// TaskHandlerFunc adapts a function to TaskHandler.
type TaskHandlerFunc func(ctx context.Context, task *asynq.Task) error

// ProcessTask calls the wrapped function.
func (f TaskHandlerFunc) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return f(ctx, task)
}

// Ensure TaskHandlerFunc implements TaskHandler
var _ TaskHandler = (TaskHandlerFunc)(nil)

// HandlerAdapter converts asynq.Handler to TaskHandler.
// Useful for integrating existing asynq handlers.
type HandlerAdapter struct {
	handler asynq.Handler
}

// NewHandlerAdapter creates a TaskHandler from asynq.Handler.
func NewHandlerAdapter(h asynq.Handler) TaskHandler {
	return &HandlerAdapter{handler: h}
}

// ProcessTask delegates to the wrapped asynq.Handler.
func (a *HandlerAdapter) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return a.handler.ProcessTask(ctx, task)
}

// Ensure HandlerAdapter implements TaskHandler
var _ TaskHandler = (*HandlerAdapter)(nil)

// BaseHandler provides common functionality for task handlers.
type BaseHandler struct {
	// Can be extended with logger, metrics, etc.
}

// NewBaseHandler creates a handler with common utilities.
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// ParsePayload unmarshals task payload into target.
// Returns error if payload cannot be unmarshaled.
func (h *BaseHandler) ParsePayload(task *asynq.Task, target interface{}) error {
	if err := sonic.Unmarshal(task.Payload(), target); err != nil {
		return fmt.Errorf("failed to parse payload for task %s: %w", task.Type(), err)
	}
	return nil
}

// ParsePayload is a helper for handlers to unmarshal task payloads.
// Use this in your handler implementations:
//
//	type EmailPayload struct { UserID int `json:"user_id"` }
//	var payload EmailPayload
//	if err := asynq.ParsePayload(t, &payload); err != nil { return err }
func ParsePayload(task *asynq.Task, target interface{}) error {
	h := &BaseHandler{}
	return h.ParsePayload(task, target)
}

// TaskError wraps errors with task context.
type TaskError struct {
	TaskType string
	Err      error
}

// Error implements error interface.
func (e *TaskError) Error() string {
	return fmt.Sprintf("task %s failed: %v", e.TaskType, e.Err)
}

// Unwrap returns the underlying error.
func (e *TaskError) Unwrap() error {
	return e.Err
}

// NewTaskError creates an error with task context.
func NewTaskError(task *asynq.Task, err error) error {
	if err == nil {
		return nil
	}
	return &TaskError{
		TaskType: task.Type(),
		Err:      err,
	}
}
