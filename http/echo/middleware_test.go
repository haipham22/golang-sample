package echo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	httpEcho "github.com/haipham22/govern/http/echo"
)

func TestWrapHandler(t *testing.T) {
	tests := []struct {
		name        string
		httpHandler http.Handler
		method      string
		path        string
		wantStatus  int
		wantBody    string
	}{
		{
			name: "wraps standard http handler successfully",
			httpHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("standard handler response"))
			}),
			method:     "GET",
			path:       "/test",
			wantStatus: http.StatusOK,
			wantBody:   "standard handler response",
		},
		{
			name: "handles POST requests",
			httpHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("created"))
			}),
			method:     "POST",
			path:       "/create",
			wantStatus: http.StatusCreated,
			wantBody:   "created",
		},
		{
			name: "preserves request headers",
			httpHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				customHeader := r.Header.Get("X-Custom-Header")
				if customHeader != "test-value" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("headers preserved"))
			}),
			method:     "GET",
			path:       "/headers",
			wantStatus: http.StatusOK,
			wantBody:   "headers preserved",
		},
		{
			name: "handles handler with error writing",
			httpHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handler that writes status but no body
				w.WriteHeader(http.StatusNoContent)
			}),
			method:     "GET",
			path:       "/no-content",
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Wrap the standard http.Handler
			wrappedHandler := httpEcho.WrapHandler(tt.httpHandler)
			e.Add(tt.method, tt.path, wrappedHandler)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			if tt.name == "preserves request headers" {
				req.Header.Set("X-Custom-Header", "test-value")
			}
			rec := httptest.NewRecorder()

			// Serve request
			e.ServeHTTP(rec, req)

			// Assert results
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestWrapHandler_WithQueryParams(t *testing.T) {
	e := echo.New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("key") != "value" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("query params ok"))
	})

	e.GET("/test", httpEcho.WrapHandler(handler))

	req := httptest.NewRequest("GET", "/test?key=value", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "query params ok", rec.Body.String())
}

func TestWrapHandler_WithRequestBody(t *testing.T) {
	e := echo.New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, err := r.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("body received"))
	})

	e.POST("/test", httpEcho.WrapHandler(handler))

	req := httptest.NewRequest("POST", "/test", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
