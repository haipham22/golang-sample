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

	apiValidator "github.com/haipham22/golang-sample/internal/validator"
	authctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
	productctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/product"
	authservice "github.com/haipham22/golang-sample/internal/usecase/auth"
	productservice "github.com/haipham22/golang-sample/internal/usecase/product"
	productRepo "github.com/haipham22/golang-sample/internal/repository/postgres"
	userRepo "github.com/haipham22/golang-sample/internal/repository/user"
	"github.com/haipham22/golang-sample/internal/orm"
)

// jwtSecretTest is a 32+ char secret for integration tests.
const jwtSecretTest = "integration-test-secret-32-chars-long"

// setupTestEngine wires the full HTTP stack (router + error handler + validator)
// over a SQLite in-memory DB so we can exercise real HTTP requests end-to-end
// without a live Postgres. Each test gets a fresh engine (and thus a fresh
// rate-limiter budget for /api).
func setupTestEngine(t *testing.T) *echo.Echo {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1) // SQLite shared cache: single conn
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&orm.User{}, &orm.Product{}))

	log := zap.NewNop().Sugar()
	storage := userRepo.New(log, db)
	svc := authservice.NewAuthService(log, storage, jwtSecretTest, 72*time.Hour)
	productStorage := productRepo.New(log, db)
	productSvc := productservice.NewService(log, productStorage)

	e := echo.New()
	e.Validator = apiValidator.NewCustomValidator()
	e.HTTPErrorHandler = makeHTTPErrorHandler(log)
	e = initRouter(e, authctrl.New(svc), healthctrl.New(db), productctrl.New(productSvc))
	return e
}

// doRequest fires an HTTP request at the engine and returns the recorder.
func doRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestIntegration_Health verifies health endpoints return 200.
func TestIntegration_Health(t *testing.T) {
	e := setupTestEngine(t)
	for _, p := range []string{"/health", "/readyz", "/livez"} {
		rec := doRequest(e, http.MethodGet, p, "")
		assert.Equal(t, http.StatusOK, rec.Code, "path %s", p)
	}
}

// TestIntegration_RegisterAndLogin covers the happy-path register → login flow.
func TestIntegration_RegisterAndLogin(t *testing.T) {
	e := setupTestEngine(t)
	regBody := `{"username":"alice","email":"alice@example.com","password":"supersecret","full_name":"Alice"}`

	rec := doRequest(e, http.MethodPost, "/api/register", regBody)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"username":"alice"`)

	// Login with the registered credentials.
	loginBody := `{"username":"alice","password":"supersecret"}`
	rec = doRequest(e, http.MethodPost, "/api/login", loginBody)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token"`)
}

// TestIntegration_RegisterDuplicate verifies the conflict (409) error path.
func TestIntegration_RegisterDuplicate(t *testing.T) {
	e := setupTestEngine(t)
	body := `{"username":"bob","email":"bob@example.com","password":"supersecret","full_name":"Bob"}`

	require.Equal(t, http.StatusCreated, doRequest(e, http.MethodPost, "/api/register", body).Code)

	rec := doRequest(e, http.MethodPost, "/api/register", body)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "Resource already exists")
}

// TestIntegration_LoginWrongPassword verifies the unauthorized (401) path.
func TestIntegration_LoginWrongPassword(t *testing.T) {
	e := setupTestEngine(t)
	reg := `{"username":"carol","email":"carol@example.com","password":"correctpass","full_name":"Carol"}`
	require.Equal(t, http.StatusCreated, doRequest(e, http.MethodPost, "/api/register", reg).Code)

	rec := doRequest(e, http.MethodPost, "/api/login", `{"username":"carol","password":"wrongpass"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unauthorized")
}

// TestIntegration_InvalidJSON verifies the invalid-input (400) error path.
func TestIntegration_InvalidJSON(t *testing.T) {
	e := setupTestEngine(t)
	rec := doRequest(e, http.MethodPost, "/api/register", `{not-json`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestIntegration_NotFoundRoute verifies an unknown route yields 404 (echo.HTTPError path).
func TestIntegration_NotFoundRoute(t *testing.T) {
	e := setupTestEngine(t)
	rec := doRequest(e, http.MethodGet, "/no-such-route", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestIntegration_ProductsCRUD exercises the product vertical end-to-end
// (HTTP -> controller -> usecase -> repository -> SQLite): create, get, list,
// delete, and the not-found path.
func TestIntegration_ProductsCRUD(t *testing.T) {
	e := setupTestEngine(t)

	// Create.
	body := `{"name":"Widget","price":9.99}`
	rec := doRequest(e, http.MethodPost, "/api/products", body)
	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"Widget"`)

	// List (one item present).
	rec = doRequest(e, http.MethodGet, "/api/products", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"Widget"`)

	// Get by id (first product -> id 1).
	rec = doRequest(e, http.MethodGet, "/api/products/1", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)

	// Delete -> 204.
	rec = doRequest(e, http.MethodDelete, "/api/products/1", "")
	require.Equal(t, http.StatusNoContent, rec.Code)

	// Get after delete -> 404.
	rec = doRequest(e, http.MethodGet, "/api/products/1", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestIntegration_ProductsInvalidInput verifies validation rejects a negative
// price via the centralized error handler (400 path).
func TestIntegration_ProductsInvalidInput(t *testing.T) {
	e := setupTestEngine(t)
	rec := doRequest(e, http.MethodPost, "/api/products", `{"name":"X","price":-1}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
