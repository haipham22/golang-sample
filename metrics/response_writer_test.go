package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriter(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.status)

	n, err := rw.Write([]byte("test"))
	requireNoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, rw.written)

	n, err = rw.Write([]byte(" data"))
	requireNoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 9, rw.written)
}

func TestResponseWriterInterface(t *testing.T) {
	var _ http.ResponseWriter = &responseWriter{}

	rw := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	rw.Header().Set("Content-Type", "text/plain")
	assert.Equal(t, "text/plain", rw.Header().Get("Content-Type"))
}
