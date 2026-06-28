package apperrors

import "testing"

// TestHelpers verifies each convenience constructor assigns the right code.
func TestHelpers(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want Code
	}{
		{"invalid input", InvalidInput("bad"), CodeInvalid},
		{"not found", NotFound("user"), CodeNotFound},
		{"unauthorized", Unauthorized("no token"), CodeUnauthorized},
		{"forbidden", Forbidden("no access"), CodeForbidden},
		{"conflict", Conflict("user"), CodeConflict},
		{"internal", Internal("boom"), CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.want {
				t.Errorf("Code = %s, want %s", tt.err.Code, tt.want)
			}
			if tt.err.Error() == "" {
				t.Error("Error() should be non-empty")
			}
		})
	}
}

// TestResolve verifies HTTP status and body mapping for coded and uncoded errors.
func TestResolve(t *testing.T) {
	t.Run("coded error", func(t *testing.T) {
		status, body := Resolve(NewCode(CodeConflict, "dup"), "/api/x", "req-1")
		if status != 409 {
			t.Errorf("status = %d, want 409", status)
		}
		if body.Path != "/api/x" || body.RequestID != "req-1" {
			t.Errorf("body path/req mismatch: %+v", body)
		}
		if body.Msg == "" || body.Error != body.Msg {
			t.Error("Msg and Error should match and be non-empty")
		}
	})

	t.Run("validation can enrich errors", func(t *testing.T) {
		status, body := Resolve(NewCode(CodeInvalid, "bad"), "/api/x", "")
		body.Errors = []FieldError{{Property: "email", Msg: "required"}}
		body.Msg = "email is required"
		if status != 400 {
			t.Errorf("status = %d, want 400", status)
		}
		if len(body.Errors) != 1 || body.Errors[0].Property != "email" {
			t.Errorf("field error not enriched: %+v", body.Errors)
		}
	})

	t.Run("uncoded error defaults to 500", func(t *testing.T) {
		status, body := Resolve(nil, "/api/x", "")
		if status != 500 {
			t.Errorf("status = %d, want 500", status)
		}
		if body.Msg != "Internal Server Error" {
			t.Errorf("Msg = %q, want Internal Server Error", body.Msg)
		}
	})
}

// TestClientMessage verifies sanitized messages per code (no leaks on 5xx).
func TestClientMessage(t *testing.T) {
	if got := CodeInternal.ClientMessage(); got != "Internal Server Error" {
		t.Errorf("internal msg = %q", got)
	}
	if got := Code("UNKNOWN").ClientMessage(); got != "Internal Server Error" {
		t.Errorf("unknown msg = %q, want Internal Server Error", got)
	}
	if got := CodeConflict.ClientMessage(); got != "Resource already exists" {
		t.Errorf("conflict msg = %q", got)
	}
}
