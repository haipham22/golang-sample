package apperrors

import (
	stderrors "errors"
	"testing"

	"go.uber.org/zap"
)

// TestLogRequestError exercises all branches without panicking. Uses a nop
// logger; the test asserts only that the function handles each case cleanly
// (nil logger, conflict, 5xx, 4xx).
func TestLogRequestError(t *testing.T) {
	log := zap.NewNop().Sugar()

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"conflict warns", NewCode(CodeConflict, "dup"), 409},
		{"5xx errors", NewCode(CodeInternal, "boom"), 500},
		{"4xx warns", NewCode(CodeNotFound, "missing"), 404},
		{"plain error", stderrors.New("plain"), 500},
		{"nil error", nil, 500},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Must not panic.
			LogRequestError(log, c.err, "/api/x", c.status)
		})
	}

	t.Run("nil logger is safe", func(t *testing.T) {
		LogRequestError(nil, stderrors.New("x"), "/p", 500)
	})
}
