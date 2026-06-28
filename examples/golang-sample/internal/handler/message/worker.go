// Package message wires asynq (govern/mq/asynq) task handlers. It registers
// sample task types (e.g. "email:send") on a *asynq.TaskMux and exposes them
// as govern/mq/asynq.TaskHandler implementations. The mux is what a caller
// passes to asynq.NewServer(redisClient, mux).
package message

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	governasynq "github.com/haipham22/govern/mq/asynq"
)

// Well-known task types registered by this package. Keep them namespaced by
// domain (":") to match asynq conventions.
const (
	TaskEmailSend = "email:send"
)

// EmailPayload is the payload schema for TaskEmailSend tasks.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Registrar builds a *asynq.TaskMux with all sample handlers registered.
type Registrar struct {
	log *zap.SugaredLogger
}

// New creates a Registrar wired with a logger.
func New(log *zap.SugaredLogger) *Registrar {
	return &Registrar{log: log}
}

// Mux returns a *asynq.TaskMux with every sample handler registered. Panics on
// duplicate registration (programming error), matching asynq.TaskMux semantics.
func (r *Registrar) Mux() *governasynq.TaskMux {
	mux := governasynq.NewTaskMux(governasynq.WithMuxLogger(r.log))
	mux.Handle(TaskEmailSend, NewEmailHandler(r.log))
	return mux
}

// EmailHandler processes email:send tasks by parsing the payload and logging
// the send (a real implementation would call an email provider).
type EmailHandler struct {
	log *zap.SugaredLogger
}

// Compile-time guard: EmailHandler satisfies govern/mq/asynq.TaskHandler.
var _ governasynq.TaskHandler = (*EmailHandler)(nil)

// NewEmailHandler creates an email:send handler.
func NewEmailHandler(log *zap.SugaredLogger) *EmailHandler {
	return &EmailHandler{log: log}
}

// ProcessTask parses the email payload and "sends" it. Returns an error when
// the payload is malformed so asynq retries per its retry policy.
func (h *EmailHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task == nil {
		return fmt.Errorf("email handler: nil task")
	}
	var payload EmailPayload
	if err := governasynq.ParsePayload(task, &payload); err != nil {
		return fmt.Errorf("email handler: parse payload: %w", err)
	}
	if payload.To == "" {
		return fmt.Errorf("email handler: payload.to is required")
	}

	// Template: replace with a real email provider call (SES, SendGrid, ...).
	h.log.Infow("sending email",
		"to", payload.To,
		"subject", payload.Subject,
		"task_type", task.Type(),
	)
	return nil
}
