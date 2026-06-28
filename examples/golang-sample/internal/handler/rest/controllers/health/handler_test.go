package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
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

func TestHTTPHandler_Check_DatabaseErrors(t *testing.T) {
	t.Run("returns degraded when DB() fails", func(t *testing.T) {
		// Create a closed DB that will fail on DB() call
		db := mockDB(t)
		sqlDB, _ := db.DB()
		sqlDB.Close() // Close the connection

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := &Controller{db: db}
		err := handler.Check(c)

		// Should return error (degraded status)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, `"status":"degraded"`)
		assert.Contains(t, body, `"database":"error"`)
	})

	t.Run("returns degraded when Ping() fails", func(t *testing.T) {
		// Use a mock DB that fails on Ping
		// Since we can't easily mock Ping(), we'll close the DB which makes Ping fail
		db := mockDB(t)
		sqlDB, _ := db.DB()
		sqlDB.Close() // Closing makes Ping fail

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := &Controller{db: db}
		_ = handler.Check(c)

		// Should return degraded status
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, `"status":"degraded"`)
		assert.Contains(t, body, `"database":"error"`)
	})

	t.Run("returns healthy when database is accessible", func(t *testing.T) {
		db := mockDB(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := &Controller{db: db}
		err := handler.Check(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, `"status":"ok"`)
		assert.Contains(t, body, `"database":"ok"`)
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
