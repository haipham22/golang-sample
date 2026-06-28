package rest

import (
	"fmt"
	"net/http"
	"testing"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/labstack/echo/v5"
)

// BenchmarkResolveError benchmarks error resolution across different error types
// This is a lightweight benchmark that tests the core error handling logic
func BenchmarkResolveError(b *testing.B) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "invalid_input",
			err:  apperrors.InvalidInput("validation failed"),
		},
		{
			name: "not_found",
			err:  apperrors.NotFound("resource"),
		},
		{
			name: "conflict",
			err:  apperrors.Conflict("resource"),
		},
		{
			name: "unauthorized",
			err:  apperrors.Unauthorized("authentication required"),
		},
		{
			name: "forbidden",
			err:  apperrors.Forbidden("access denied"),
		},
		{
			name: "internal",
			err:  apperrors.Internal("something went wrong"),
		},
		{
			name: "echo_http_error",
			err:  echo.NewHTTPError(http.StatusBadRequest, "bad request"),
		},
		{
			name: "generic_error",
			err:  fmt.Errorf("generic error"),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = resolveError(tc.err, "/api/test", "req-123")
			}
		})
	}
}
