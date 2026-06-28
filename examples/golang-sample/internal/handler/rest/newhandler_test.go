package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/pkg/config"

	authctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
	productctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/product"
	"github.com/haipham22/golang-sample/internal/orm"
	productRepo "github.com/haipham22/golang-sample/internal/repository/postgres"
	userRepo "github.com/haipham22/golang-sample/internal/repository/user"
	authservice "github.com/haipham22/golang-sample/internal/usecase/auth"
	productservice "github.com/haipham22/golang-sample/internal/usecase/product"
)

// buildEngineWithNewHandler wires NewHandler against an in-memory SQLite DB.
// NewHandler is otherwise only reachable via rest.New, which needs a live
// Postgres; this factory lets handler.go's middleware + router wiring be
// exercised end-to-end without one. The returned Echo has all middleware,
// routes, validator, and error handler installed by NewHandler.
func buildEngineWithNewHandler(t *testing.T) *echo.Echo {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&orm.User{}, &orm.Product{}))

	log := zap.NewNop().Sugar()
	storage := userRepo.New(log, db)
	authSvc := authservice.NewAuthService(log, storage, jwtSecretTest, 72*time.Hour)
	productStorage := productRepo.New(log, db)
	productSvc := productservice.NewService(log, productStorage)

	e := echo.New()
	// NewHandler mutates e in place (middleware, validator, error handler,
	// routes) and returns a server bound to it. We discard the server — its
	// *echo.Echo is the same instance we just built, which is what we test.
	_ = NewHandler(
		log,
		e,
		authctrl.New(authSvc),
		healthctrl.New(db),
		productctrl.New(productSvc),
		0,
		true,
		config.EnvDevelopment,
	)
	return e
}

// TestNewHandler_WiresRoutesAndMiddleware drives NewHandler end-to-end and
// asserts routes are registered, SecurityHeaders middleware is applied, and the
// custom error handler is installed. NewHandler is otherwise 0% covered
// (reachable only through rest.New, which requires Postgres).
func TestNewHandler_WiresRoutesAndMiddleware(t *testing.T) {
	e := buildEngineWithNewHandler(t)

	cases := []struct {
		method, path string
		body         string
		wantStatus   int
	}{
		{http.MethodGet, "/health", "", http.StatusOK},
		{http.MethodGet, "/readyz", "", http.StatusOK},
		{http.MethodGet, "/livez", "", http.StatusOK},
		{http.MethodGet, "/api/products", "", http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, c.wantStatus, rec.Code, "path %s", c.path)
			// SecurityHeaders middleware runs on every response — proves it was wired.
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		})
	}
}

// TestNewHandler_SwaggerRouteRegistered verifies the Swagger gate
// (debug && non-production) registers /docs/* when debug is true.
func TestNewHandler_SwaggerRouteRegistered(t *testing.T) {
	e := buildEngineWithNewHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "Swagger docs route should be registered")
}

// TestNewHandler_ErrorHandlerSanitizes5xx verifies NewHandler installed
// makeHTTPErrorHandler by triggering a 5xx echo.HTTPError and asserting the
// sanitized message.
func TestNewHandler_ErrorHandlerSanitizes5xx(t *testing.T) {
	e := buildEngineWithNewHandler(t)

	e.GET("/_test/boom", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "secret-db-creds-leaked")
	})

	req := httptest.NewRequest(http.MethodGet, "/_test/boom", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret-db-creds-leaked")
	assert.Contains(t, rec.Body.String(), "Internal Server Error")
}

// TestNewHandler_TrailingSlashRedirect verifies the RemoveTrailingSlash
// middleware (configured for 308) is wired.
func TestNewHandler_TrailingSlashRedirect(t *testing.T) {
	e := buildEngineWithNewHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/health/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
	assert.Equal(t, "/health", rec.Header().Get(echo.HeaderLocation))
}
