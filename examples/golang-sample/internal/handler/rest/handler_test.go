package rest

import (
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	schemas "github.com/haipham22/golang-sample/internal/schemas"
	apiValidator "github.com/haipham22/golang-sample/internal/validator"
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

// TestResolveError_ValidationEnrichment verifies CodeInvalid errors wrapping a
// validator.ValidationError get field-level details.
func TestResolveError_ValidationEnrichment(t *testing.T) {
	ve := &apiValidator.ValidationError{Detail: schemas.ErrorDetail{
		Property: "email",
		Msg:      "email is required",
	}}
	err := apperrors.WrapCode(apperrors.CodeInvalid, ve)

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
