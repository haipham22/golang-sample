package rest

import (
	"context"

	"golang-sample/internal/handler/rest/controllers/auth"
	"golang-sample/internal/handler/rest/controllers/health"
	"golang-sample/internal/handler/rest/middlewares"

	"github.com/labstack/echo/v4"
)

func initRouter(
	e *echo.Echo,
	authCtrl *auth.Controller,
	healthCtrl *health.Controller,
) *echo.Echo {
	// Health check endpoints
	e.GET("/health", healthCtrl.Check)
	e.GET("/readyz", healthCtrl.Ready)
	e.GET("/livez", healthCtrl.Live)

	public := e.Group("/api")

	// Apply rate limiting to auth endpoints (10 requests per minute per IP)
	// Using context.Background() as the rate limiter runs for the application lifetime
	public := e.Group("/api")
	authRateLimiter := middlewares.RateLimit(context.Background())
	public.POST("/login", authCtrl.PostLogin, authRateLimiter)
	public.POST("/register", authCtrl.PostRegister, authRateLimiter)

	return e
}
