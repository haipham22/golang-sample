package internal

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"

	"golang-sample/pkg/config"
)

func (h *Handler) setRoutes() error {
	// Health checks (no auth required - for Kubernetes probes)
	h.server.GET("/health", h.health.Check)
	h.server.GET("/readyz", h.health.Ready)
	h.server.GET("/livez", h.health.Live)

	// Metrics endpoint (no auth required - for Prometheus scraping)
	h.server.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	if config.ENV.APP.ENV != config.EnvProduction {
		h.server.GET("/document/*", echoSwagger.WrapHandler)
	}

	public := h.server.Group("/api")

	public.POST("/login", h.auth.PostLogin)
	public.POST("/register", h.auth.PostRegister)

	return nil
}
