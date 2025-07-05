package api

import (
	echoSwagger "github.com/swaggo/echo-swagger"

	"golang-sample/pkg/config"
)

func (h *Handler) setRoutes() error {
	if config.ENV.APP.ENV != config.EnvProduction {
		h.server.GET("/document/*", echoSwagger.WrapHandler)
	}

	public := h.server.Group("/api")

	public.POST("/login", h.auth.PostLogin)
	public.POST("/register", h.auth.PostRegister)

	return nil
}
