package apperrors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

// TestErrorFormatting covers all branches of (*Error).Error().
func TestErrorFormatting(t *testing.T) {
	base := stderrors.New("db down")
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{"message+cause", &Error{Code: CodeInternal, message: "create user", Err: base}, "create user: db down"},
		{"message only", &Error{Code: CodeInvalid, message: "bad email"}, "bad email"},
		{"cause only", &Error{Code: CodeInternal, Err: base}, "db down"},
		{"code only", &Error{Code: CodeConflict}, string(CodeConflict)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNilErrorSafe ensures nil receivers do not panic.
func TestNilErrorSafe(t *testing.T) {
	var e *Error
	if got := e.Error(); got != "<nil>" {
		t.Errorf("nil Error() = %q, want <nil>", got)
	}
	if e.Unwrap() != nil {
		t.Error("nil Unwrap() should be nil")
	}
	if e.Message() != "" {
		t.Error("nil Message() should be empty")
	}
}

// TestUnwrapChain verifies stderrors.Is/As traverse wrapped apperrors.
func TestUnwrapChain(t *testing.T) {
	sentinel := stderrors.New("root cause")
	wrapped := Wrap(CodeInternal, sentinel)

	if !stderrors.Is(wrapped, sentinel) {
		t.Error("stderrors.Is should find wrapped sentinel")
	}

	var target *Error
	if !stderrors.As(wrapped, &target) {
		t.Fatal("stderrors.As should find *Error")
	}
	if target.Code != CodeInternal {
		t.Errorf("target code = %s, want %s", target.Code, CodeInternal)
	}
}

// TestGetCode covers code extraction including through wrapping layers.
func TestGetCode(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		c, ok := GetCode(NewCode(CodeConflict, "dup"))
		if !ok || c != CodeConflict {
			t.Errorf("GetCode = (%s,%v), want (%s,true)", c, ok, CodeConflict)
		}
	})
	t.Run("wrapped by fmt", func(t *testing.T) {
		outer := fmt.Errorf("outer: %w", NewCode(CodeNotFound, "missing"))
		c, ok := GetCode(outer)
		if !ok || c != CodeNotFound {
			t.Errorf("GetCode through fmt.Errorf = (%s,%v), want (%s,true)", c, ok, CodeNotFound)
		}
	})
	t.Run("no code", func(t *testing.T) {
		c, ok := GetCode(stderrors.New("plain"))
		if ok {
			t.Errorf("GetCode on plain error = (%s,true), want false", c)
		}
	})
	t.Run("nil", func(t *testing.T) {
		if _, ok := GetCode(nil); ok {
			t.Error("GetCode(nil) should be false")
		}
	})
}

// TestIsCode verifies IsCode matching.
func TestIsCode(t *testing.T) {
	if !IsCode(Wrap(CodeUnauthorized, stderrors.New("no token")), CodeUnauthorized) {
		t.Error("IsCode should match wrapped code")
	}
	if IsCode(stderrors.New("plain"), CodeInternal) {
		t.Error("IsCode should be false for plain errors")
	}
}

// TestWrapCodeNil verifies WrapCode(nil) returns nil (prevents phantom errors).
func TestWrapCodeNil(t *testing.T) {
	if err := WrapCode(CodeInternal, nil); err != nil {
		t.Errorf("WrapCode(_, nil) = %v, want nil", err)
	}
}

// TestSentinelsAreRecognized ensures sentinels are *Error and GetCode works.
func TestSentinelsAreRecognized(t *testing.T) {
	tests := []struct {
		err  error
		want Code
	}{
		{ErrInternal, CodeInternal},
		{ErrInvalid, CodeInvalid},
		{ErrNotFound, CodeNotFound},
		{ErrUnauthorized, CodeUnauthorized},
	}
	for _, tt := range tests {
		c, ok := GetCode(tt.err)
		if !ok || c != tt.want {
			t.Errorf("GetCode(%v) = (%s,%v), want (%s,true)", tt.err, c, ok, tt.want)
		}
		// stderrors.Is must match a wrapped sentinel.
		wrapped := fmt.Errorf("ctx: %w", tt.err)
		if !stderrors.Is(wrapped, tt.err) {
			t.Errorf("stderrors.Is should match wrapped sentinel %v", tt.err)
		}
	}
}

// TestHTTPStatusMapping verifies canonical HTTP status per code.
func TestHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		code    Code
		wantSta int
	}{
		{CodeInvalid, 400},
		{CodeUnauthorized, 401},
		{CodeForbidden, 403},
		{CodeNotFound, 404},
		{CodeConflict, 409},
		{CodeAlreadyExists, 409},
		{CodeRateLimit, 429},
		{CodeInternal, 500},
		{Code("UNKNOWN"), 500}, // unknown defaults to 500
	}
	for _, tt := range tests {
		if got := tt.code.HTTPStatus(); got != tt.wantSta {
			t.Errorf("%s.HTTPStatus() = %d, want %d", tt.code, got, tt.wantSta)
		}
	}
}
