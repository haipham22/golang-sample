package rest

import (
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// TestResolveError_AppErrors verifies centralized code -> status/body mapping.
func TestResolveError_AppErrors(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantStat int
		wantMsg  string
	}{
		{"invalid", apperrors.NewCode(apperrors.CodeInvalid, "bad"), http.StatusBadRequest, "invalid request parameters"},
		{"not found", apperrors.NewCode(apperrors.CodeNotFound, "x"), http.StatusNotFound, "Resource not found"},
		{"unauthorized", apperrors.ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
		{"forbidden", apperrors.NewCode(apperrors.CodeForbidden, "x"), http.StatusForbidden, "Forbidden"},
		{"conflict", apperrors.NewCode(apperrors.CodeConflict, "x"), http.StatusConflict, "Resource already exists"},
		{"internal", apperrors.NewCode(apperrors.CodeInternal, "x"), http.StatusInternalServerError, "Internal Server Error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body := resolveError(c.err, "/api/x", "req-1")
			if status != c.wantStat {
				t.Errorf("status = %d, want %d", status, c.wantStat)
			}
			if body.Msg != c.wantMsg || body.Error != c.wantMsg {
				t.Errorf("body msg/error = %q/%q, want %q", body.Msg, body.Error, c.wantMsg)
			}
			if body.Path != "/api/x" || body.RequestID != "req-1" {
				t.Errorf("body path/req = %q/%q", body.Path, body.RequestID)
			}
		})
	}
}

// TestResolveError_ValidationEnrichment verifies CodeInvalid errors with field
// detail are resolved by internal/errors, not by handler-specific enrichment.
func TestResolveError_ValidationEnrichment(t *testing.T) {
	err := apperrors.Validation("email", "email is required")

	status, body := resolveError(err, "/api/x", "")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if body.Msg != "email is required" {
		t.Errorf("Msg = %q, want field message", body.Msg)
	}
	if len(body.Errors) != 1 || body.Errors[0].Property != "email" {
		t.Errorf("field errors = %+v, want email", body.Errors)
	}
}

// TestResolveError_EchoHTTPError verifies echo.HTTPError pass-through + 5xx
// sanitization.
func TestResolveError_EchoHTTPError(t *testing.T) {
	t.Run("4xx passes message", func(t *testing.T) {
		status, body := resolveError(echo.NewHTTPError(http.StatusNotFound, "nope"), "/p", "")
		if status != http.StatusNotFound || body.Msg != "nope" {
			t.Errorf("got %d %q", status, body.Msg)
		}
	})
	t.Run("5xx sanitized", func(t *testing.T) {
		status, body := resolveError(echo.NewHTTPError(http.StatusInternalServerError, "db credentials leaked"), "/p", "")
		if status != http.StatusInternalServerError {
			t.Fatalf("status = %d", status)
		}
		if body.Msg == "db credentials leaked" {
			t.Error("5xx message must be sanitized, not leak internal details")
		}
	})
}

// TestResolveError_UnknownDefaults500 verifies plain errors map to 500.
func TestResolveError_UnknownDefaults500(t *testing.T) {
	status, body := resolveError(stderrors.New("boom"), "/p", "")
	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", status)
	}
	if body.Msg != "Internal Server Error" {
		t.Errorf("Msg = %q", body.Msg)
	}
}

// httpStatusCoderErr is a non-*echo.HTTPError type that implements Echo's
// HTTPStatusCoder interface, letting us exercise the echo.StatusCode sentinel
// branch of resolveError (which mirrors how Echo's own unexported httpError
// sentinels like ErrInternal are surfaced).
type httpStatusCoderErr struct{ code int }

func (e httpStatusCoderErr) Error() string   { return "coder error" }
func (e httpStatusCoderErr) StatusCode() int { return e.code }

// TestResolveError_EchoStatusCodeSentinel covers the branch where the error is
// not *echo.HTTPError but implements HTTPStatusCoder (echo.StatusCode != 0),
// plus the 5xx sanitization inside that branch.
func TestResolveError_EchoStatusCodeSentinel(t *testing.T) {
	t.Run("4xx passes StatusText", func(t *testing.T) {
		status, body := resolveError(httpStatusCoderErr{code: http.StatusNotFound}, "/p", "r")
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", status)
		}
		if body.Msg != "Not Found" {
			t.Errorf("Msg = %q, want %q", body.Msg, http.StatusText(http.StatusNotFound))
		}
	})

	t.Run("5xx sanitized", func(t *testing.T) {
		status, body := resolveError(httpStatusCoderErr{code: http.StatusServiceUnavailable}, "/p", "r")
		if status != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", status)
		}
		if body.Msg != "Internal Server Error" {
			t.Errorf("5xx Msg = %q, want sanitized %q", body.Msg, "Internal Server Error")
		}
	})
}

// TestResolveError_EchoHTTPError_EmptyMessage covers the empty-message fallback
// (clientMsg == "" -> http.StatusText(code)) in the *echo.HTTPError branch.
func TestResolveError_EchoHTTPError_EmptyMessage(t *testing.T) {
	status, body := resolveError(echo.NewHTTPError(http.StatusNotFound, ""), "/p", "")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if body.Msg != "Not Found" {
		t.Errorf("Msg = %q, want fallback %q", body.Msg, "Not Found")
	}
}
