package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockDB creates a test database connection
func mockDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	return db
}

func TestHTTPHandler_Check(t *testing.T) {
	t.Run("healthy with valid DB", func(t *testing.T) {
		// Setup
		db := mockDB(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create handler
		handler := &Controller{db: db}

		// Execute
		err := handler.Check(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check response contains expected fields
		assert.Contains(t, rec.Body.String(), `"status":"ok"`)
		assert.Contains(t, rec.Body.String(), `"database":"ok"`)
		assert.Contains(t, rec.Body.String(), `"service":"golang-sample-api"`)
		assert.Contains(t, rec.Body.String(), `"timestamp"`)
	})
}

func TestHTTPHandler_Ready(t *testing.T) {
	t.Run("always returns ready", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create handler (DB doesn't matter for Ready)
		handler := &Controller{db: nil}

		// Execute
		err := handler.Ready(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"ready"}`, rec.Body.String())
	})
}

func TestHTTPHandler_Live(t *testing.T) {
	t.Run("always returns alive", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create handler (DB doesn't matter for Live)
		handler := &Controller{db: nil}

		// Execute
		err := handler.Live(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"alive"}`, rec.Body.String())
	})
}

func TestNewHTTPHandler(t *testing.T) {
	db := mockDB(t)

	handler := New(db)

	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

func TestHTTPHandler_WithDatabase(t *testing.T) {
	t.Run("handler with valid DB", func(t *testing.T) {
		db := mockDB(t)
		handler := &Controller{db: db}

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.db)
	})

	t.Run("handler with nil DB", func(t *testing.T) {
		handler := &Controller{db: nil}

		assert.NotNil(t, handler)
		assert.Nil(t, handler.db)
	})
}

// Benchmark tests
func BenchmarkHTTPHandler_Check(b *testing.B) {
	db := mockDB(&testing.T{})
	handler := &Controller{db: db}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = handler.Check(c)
	}
}

func BenchmarkHTTPHandler_Ready(b *testing.B) {
	handler := &Controller{db: nil}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = handler.Ready(c)
	}
}

func BenchmarkHTTPHandler_Live(b *testing.B) {
	handler := &Controller{db: nil}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = handler.Live(c)
	}
}
