package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPMiddlewareStreaming(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		for i := 0; i < 3; i++ {
			_, _ = w.Write([]byte("chunk"))
		}
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/stream", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "chunkchunkchunk", rec.Body.String())
}

func TestHTTPMiddlewareWriteString(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "hello via io.WriteString")
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello via io.WriteString", rec.Body.String())
}

func TestHTTPMiddlewareNoWrite(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("DELETE", "/resource", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assertEmpty(t, rec.Body.String())
}

func TestMiddlewareRecordsMultipleWrites(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 1000)))
		_, _ = w.Write([]byte(strings.Repeat("b", 2000)))
		_, _ = w.Write([]byte(strings.Repeat("c", 3000)))
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, 6000, rec.Body.Len())
}

func TestMiddlewareHandlesEmptyBody(t *testing.T) {
	reg := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	wrapped := HTTPMiddlewareWithRegistry(reg, handler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assertEqual(t, 0, rec.Body.Len())
}
