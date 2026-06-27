package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPMiddleware(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler, "service")
	requireNotNil(t, wrapped)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestHTTPMiddlewareSuccess(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler, "service")

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHTTPMiddlewareStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{
			name:           "default OK",
			handler:        func(w http.ResponseWriter, r *http.Request) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "explicit OK",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/elsewhere", http.StatusFound)
			},
			expectedStatus: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := New()
			wrapped := HTTPMiddlewareWithRegistry(reg, tt.handler)
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestHTTPMiddlewareStatusText(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("I'm a teapot"))
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/teapot", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Contains(t, rec.Body.String(), "teapot")
}

func TestHTTPMiddlewareWithPanic(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	rec := httptest.NewRecorder()

	assert.Panics(t, func() {
		wrapped.ServeHTTP(rec, req)
	})
}

func TestHTTPMiddlewareCustomLabels(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler, "service", "version")
	requireNotNil(t, wrapped)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
