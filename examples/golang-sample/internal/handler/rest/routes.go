package rest

import (
	"context"

	"github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
	"github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
	"github.com/haipham22/golang-sample/internal/handler/rest/controllers/product"
	"github.com/haipham22/golang-sample/internal/handler/rest/middlewares"

	"github.com/labstack/echo/v4"
)

func initRouter(
	e *echo.Echo,
	authCtrl *auth.Controller,
	healthCtrl *health.Controller,
	productCtrl *product.Controller,
) *echo.Echo {
	// Health check endpoints
	e.GET("/health", healthCtrl.Check)
	e.GET("/readyz", healthCtrl.Ready)
	e.GET("/livez", healthCtrl.Live)

	public := e.Group("/api")

	// Apply rate limiting to auth endpoints (10 requests per minute per IP)
	// Using context.Background() as the rate limiter runs for the application lifetime
	authRateLimiter := middlewares.RateLimit(context.Background())
	public.POST("/login", authCtrl.PostLogin, authRateLimiter)
	public.POST("/register", authCtrl.PostRegister, authRateLimiter)

	// Product CRUD endpoints.
	products := public.Group("/products")
	products.POST("", productCtrl.PostProduct)
	products.GET("", productCtrl.ListProducts)
	products.GET("/:id", productCtrl.GetProduct)
	products.DELETE("/:id", productCtrl.DeleteProduct)

	return e
}
