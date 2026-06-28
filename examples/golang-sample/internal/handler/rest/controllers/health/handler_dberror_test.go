package health

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// nonDBConnPool is a gorm.ConnPool that is NOT a *sql.DB and does NOT implement
// gorm's GetDBConnector interface. When gorm's DB() type-asserts its ConnPool
// to *sql.DB (and to GetDBConnector), both assertions fail and DB() returns
// gorm.ErrInvalidDB — the path health.Check exercises when h.db.DB() returns
// an error. The four ConnPool methods are never invoked in our test path.
type nonDBConnPool struct{}

func (nonDBConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, nil
}
func (nonDBConnPool) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, nil
}
func (nonDBConnPool) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, nil
}
func (nonDBConnPool) QueryRowContext(context.Context, string, ...any) *sql.Row { return nil }

// TestCheck_DBMethodReturnsError covers the previously uncovered branch where
// h.db.DB() returns an error. We open a real sqlite *gorm.DB (fully
// initialized), then point the EXPORTED db.Statement.ConnPool field at our
// nonDBConnPool. gorm's DB() checks Statement.ConnPool first, sees it is not a
// *sql.DB nor a GetDBConnector, and returns ErrInvalidDB. The handler must
// respond 503 with status=degraded and database=error.
//
// No production code is altered; only the EXPORTED Statement.ConnPool field is
// reassigned.
func TestCheck_DBMethodReturnsError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// Sanity: a freshly-opened sqlite DB has a usable Statement.ConnPool.
	require.NotNil(t, db.Statement, "gorm.DB.Statement must be non-nil after Open")

	// Swap the EXPORTED Statement.ConnPool field to our stub. gorm.DB()
	// prefers Statement.ConnPool over the root db.ConnPool, so DB() now fails.
	db.Statement.ConnPool = nonDBConnPool{}

	_, dbErr := db.DB()
	require.Error(t, dbErr, "DB() must error after swapping Statement.ConnPool")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := &Controller{db: db}
	err = handler.Check(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"status":"degraded"`)
	assert.Contains(t, body, `"database":"error"`)
	assert.Contains(t, body, "Failed to get database connection")
}
