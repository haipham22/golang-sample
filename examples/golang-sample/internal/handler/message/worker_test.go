package message

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newEmailTask(t *testing.T, payload EmailPayload) *asynq.Task {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return asynq.NewTask(TaskEmailSend, data)
}

func TestEmailHandler_ProcessTask_Success(t *testing.T) {
	h := NewEmailHandler(zap.NewNop().Sugar())
	task := newEmailTask(t, EmailPayload{To: "user@example.com", Subject: "hi", Body: "hello"})

	require.NoError(t, h.ProcessTask(context.Background(), task))
}

func TestEmailHandler_ProcessTask_Errors(t *testing.T) {
	h := NewEmailHandler(zap.NewNop().Sugar())

	t.Run("nil task", func(t *testing.T) {
		err := h.ProcessTask(context.Background(), nil)
		require.Error(t, err)
	})

	t.Run("malformed payload", func(t *testing.T) {
		bad := asynq.NewTask(TaskEmailSend, []byte("{not json"))
		err := h.ProcessTask(context.Background(), bad)
		require.Error(t, err)
	})

	t.Run("missing to", func(t *testing.T) {
		task := newEmailTask(t, EmailPayload{To: "", Subject: "x"})
		err := h.ProcessTask(context.Background(), task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "to")
	})
}

func TestEmailHandler_InterfaceCompliance(t *testing.T) {
	// Compile-time: EmailHandler must satisfy govern asynq.TaskHandler.
	var _ emailHandlerLike = (*EmailHandler)(nil)
}

// emailHandlerLike mirrors govern/mq/asynq.TaskHandler so we can assert the
// contract without importing the govern interface again here.
type emailHandlerLike interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

func TestRegistrar_Mux(t *testing.T) {
	r := New(zap.NewNop().Sugar())
	mux := r.Mux()
	require.NotNil(t, mux)

	// The email handler must be registered and routable.
	assert.True(t, mux.HasHandler(TaskEmailSend))

	// Routing an unknown task type must error.
	err := mux.HandleTask(context.Background(), asynq.NewTask("unknown:type", nil))
	require.Error(t, err)
}

func TestRegistrar_Mux_RoutesEmailTask(t *testing.T) {
	r := New(zap.NewNop().Sugar())
	mux := r.Mux()

	// End-to-end through the mux: build a real task and let the mux dispatch.
	task := newEmailTask(t, EmailPayload{To: "to@example.com", Subject: "s"})
	require.NoError(t, mux.HandleTask(context.Background(), task))
}
