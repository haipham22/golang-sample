package health

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Controller handles health check requests
type Controller struct {
	db *gorm.DB
}

// New creates a new health HTTP handler
func New(db *gorm.DB) *Controller {
	return &Controller{db: db}
}

// Check performs full health check including database connectivity
func (h *Controller) Check(c echo.Context) error {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "golang-sample-api",
	}

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		status["database"] = "error"
		status["status"] = "degraded"
		status["error"] = "Failed to get database connection"
		return c.JSON(http.StatusServiceUnavailable, status)
	}

	if err := sqlDB.Ping(); err != nil {
		status["database"] = "error"
		status["status"] = "degraded"
		status["error"] = "Database connection failed"
		return c.JSON(http.StatusServiceUnavailable, status)
	}

	status["database"] = "ok"
	return c.JSON(http.StatusOK, status)
}

// Ready returns readiness status for Kubernetes probes
func (h *Controller) Ready(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}

// Live returns liveness status for Kubernetes probes
func (h *Controller) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "alive",
	})
}
