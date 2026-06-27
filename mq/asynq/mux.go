package asynq

import (
	"context"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// TaskMux routes tasks to registered handlers.
// Similar to http.ServeMux but for asynq tasks.
type TaskMux struct {
	mu       sync.RWMutex
	handlers map[string]TaskHandler
	logger   *zap.SugaredLogger
}

// NewTaskMux creates a new task handler registry.
func NewTaskMux(opts ...MuxOption) *TaskMux {
	cfg := &muxConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	logger := cfg.logger
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	return &TaskMux{
		handlers: make(map[string]TaskHandler),
		logger:   logger,
	}
}

// Handle registers a handler for the given task type.
// If a handler already exists for the type, it panics.
//
// The handler will be called for each task of this type.
func (m *TaskMux) Handle(taskType string, handler TaskHandler) {
	if handler == nil {
		panic("asynq: nil handler for task type " + taskType)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.handlers[taskType]; ok {
		panic("asynq: multiple handlers for task type " + taskType)
	}

	m.handlers[taskType] = handler
	m.logger.Debugf("Registered handler for task type: %s", taskType)
}

// HandleFunc registers a function handler for the given task type.
// Convenience wrapper for Handle with TaskHandlerFunc.
func (m *TaskMux) HandleFunc(taskType string, handler func(ctx context.Context, task *asynq.Task) error) {
	if handler == nil {
		panic("asynq: nil handler function for task type " + taskType)
	}

	m.Handle(taskType, TaskHandlerFunc(handler))
}

// HandleTask processes a task by routing it to the registered handler.
// Implements asynq.Handler interface for integration with asynq server.
func (m *TaskMux) HandleTask(ctx context.Context, task *asynq.Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	m.mu.RLock()
	handler, ok := m.handlers[task.Type()]
	m.mu.RUnlock()

	if !ok || handler == nil {
		return fmt.Errorf("no handler registered for task type: %s", task.Type())
	}

	return handler.ProcessTask(ctx, task)
}

// HandlerTypes returns all registered task types.
func (m *TaskMux) HandlerTypes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]string, 0, len(m.handlers))
	for typ := range m.handlers {
		types = append(types, typ)
	}
	return types
}

// HasHandler returns true if a handler is registered for the task type.
func (m *TaskMux) HasHandler(taskType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.handlers[taskType]
	return ok
}

// Unregister removes the handler for the given task type.
// Returns true if a handler was removed.
func (m *TaskMux) Unregister(taskType string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.handlers[taskType]; ok {
		delete(m.handlers, taskType)
		m.logger.Debugf("Unregistered handler for task type: %s", taskType)
		return true
	}
	return false
}

// ServeAsynqHandler adapts TaskMux to asynq.Handler interface.
// This allows TaskMux to be used directly with asynq.NewServer.
func (m *TaskMux) ServeAsynqHandler(ctx context.Context, t *asynq.Task) error {
	return m.HandleTask(ctx, t)
}

// ProcessTask implements asynq.Handler interface.
func (m *TaskMux) ProcessTask(ctx context.Context, t *asynq.Task) error {
	return m.HandleTask(ctx, t)
}

// Ensure TaskMux implements asynq.Handler
var _ asynq.Handler = (*TaskMux)(nil)
