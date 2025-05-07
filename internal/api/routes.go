package api

import (
	"github.com/labstack/echo/v4"

	"golang-sample/internal/api/routes/auth"
)

func SetAuthRoutes(e *echo.Group, h *auth.Controller) {
	e.POST("/login", h.PostLogin)
	e.POST("/register", h.PostRegister)
}
