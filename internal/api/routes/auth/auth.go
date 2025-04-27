package auth

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"golang-sample/internal/api/storage"
)

type Auth interface {
	PostLogin(c echo.Context) error
	PostRegister(c echo.Context) error
}

type authHandler struct {
	log     *zap.SugaredLogger
	storage *storage.Storage
}

func NewAuthHandler(log *zap.SugaredLogger, storage *storage.Storage) Auth {
	return &authHandler{
		log:     log,
		storage: storage,
	}
}

func SetAuthRoutes(e *echo.Group, h Auth) {
	e.POST("/login", h.PostLogin)
	e.POST("/register", h.PostRegister)
}
