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

func TestController_Check(t *testing.T) {
	t.Run("healthy with valid DB", func(t *testing.T) {
		// Setup
		db := mockDB(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create controller
		controller := &Controller{db: db}

		// Execute
		err := controller.Check(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check response contains expected fields
		assert.Contains(t, rec.Body.String(), `"status":"ok"`)
		assert.Contains(t, rec.Body.String(), `"database":"ok"`)
		assert.Contains(t, rec.Body.String(), `"service":"golang-sample-api"`)
		assert.Contains(t, rec.Body.String(), `"timestamp"`)
	})

	t.Run("panic with nil DB", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create controller with nil DB (will panic)
		controller := &Controller{db: nil}

		// Execute and expect panic
		assert.Panics(t, func() {
			_ = controller.Check(c)
		})
	})
}

func TestController_Ready(t *testing.T) {
	t.Run("always returns ready", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create controller (DB doesn't matter for Ready)
		controller := &Controller{db: nil}

		// Execute
		err := controller.Ready(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"ready"}`, rec.Body.String())
	})
}

func TestController_Live(t *testing.T) {
	t.Run("always returns alive", func(t *testing.T) {
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create controller (DB doesn't matter for Live)
		controller := &Controller{db: nil}

		// Execute
		err := controller.Live(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"alive"}`, rec.Body.String())
	})
}

func TestNewController(t *testing.T) {
	db := mockDB(t)

	controller := NewController(db)

	assert.NotNil(t, controller)
	assert.Equal(t, db, controller.db)
}

func TestController_WithDatabase(t *testing.T) {
	t.Run("controller with valid DB", func(t *testing.T) {
		db := mockDB(t)
		controller := &Controller{db: db}

		assert.NotNil(t, controller)
		assert.NotNil(t, controller.db)
	})

	t.Run("controller with nil DB", func(t *testing.T) {
		controller := &Controller{db: nil}

		assert.NotNil(t, controller)
		assert.Nil(t, controller.db)
	})
}

// Benchmark tests
func BenchmarkController_Check(b *testing.B) {
	db := mockDB(&testing.T{})
	controller := &Controller{db: db}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = controller.Check(c)
	}
}

func BenchmarkController_Ready(b *testing.B) {
	controller := &Controller{db: nil}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = controller.Ready(c)
	}
}

func BenchmarkController_Live(b *testing.B) {
	controller := &Controller{db: nil}
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = controller.Live(c)
	}
}
