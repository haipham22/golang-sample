package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Passing(t *testing.T) {
	r := New()

	r.Register("db", func(ctx context.Context) error {
		return nil
	})

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	rec := httptest.NewRecorder()

	r.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp Response
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, StatusPassing, resp.Status)
}

func TestHandler_Failing(t *testing.T) {
	r := New()

	r.Register("db", func(ctx context.Context) error {
		return errors.New("failed")
	})

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	rec := httptest.NewRecorder()

	r.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp Response
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFailing, resp.Status)
}

func TestHandler_SingleCheck(t *testing.T) {
	r := New()

	r.Register("db", func(ctx context.Context) error {
		return errors.New("failed")
	})

	r.Register("cache", func(ctx context.Context) error {
		return nil
	})

	req := httptest.NewRequest("GET", "/health?name=cache", http.NoBody)
	rec := httptest.NewRecorder()

	r.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result Result
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "cache", result.Name)
	assert.Equal(t, StatusPassing, result.Status)
}

func TestHandler_SingleCheckNotFound(t *testing.T) {
	r := New()

	req := httptest.NewRequest("GET", "/health?name=unknown", http.NoBody)
	rec := httptest.NewRecorder()

	r.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestLiveness(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", http.NoBody)
	rec := httptest.NewRecorder()

	Liveness().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestHandler_POST(t *testing.T) {
	r := New()

	r.Register("db", func(ctx context.Context) error {
		return nil
	})

	req := httptest.NewRequest("POST", "/health", http.NoBody)
	rec := httptest.NewRecorder()

	r.Handler().ServeHTTP(rec, req)

	// Should work with any method
	assert.Equal(t, http.StatusOK, rec.Code)
}
