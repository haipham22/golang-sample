package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	"golang-sample/internal/api/middlewares"
	"golang-sample/internal/api/routes/auth"
	apiValidator "golang-sample/internal/api/validator"
	"golang-sample/pkg/config"
)

type (
	Handler struct {
		log    *zap.SugaredLogger
		server *echo.Echo

		auth *auth.Controller
	}

	ServerFunc struct {
		Start    func() error
		Shutdown func(context.Context) error
	}
)

const (
	readHeaderTimeout = 30 * time.Second
)

func NewHandler(log *zap.SugaredLogger, e *echo.Echo, auth *auth.Controller) *Handler {

	e.Validator = apiValidator.NewCustomValidator()

	if config.ENV.APP.DEBUG {
		e.Debug = true
	}

	e.Use(
		middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
			RedirectCode: http.StatusMovedPermanently,
		}),
		middleware.Recover(),
		middleware.RequestID(),
		//middleware.GzipWithConfig(middleware.GzipConfig{
		//	Skipper: func(c echo.Context) bool {
		//		return strings.Contains(c.Request().URL.Path, "document")
		//	},
		//}),
		middlewares.Logger(zap.L()),
	)

	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	if config.ENV.APP.ENV != config.EnvProduction {
		e.GET("/document/*", echoSwagger.WrapHandler)
	}

	return &Handler{
		log:    log,
		server: e,
		auth:   auth,
	}
}

func (h *Handler) CreateServer(port int64) (*ServerFunc, error) {

	if err := h.setRoutes(); err != nil {
		h.log.Fatalf("Could not set routes: %v", err)
	}

	sv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           h.server,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return &ServerFunc{
		Start: sv.ListenAndServe,
		Shutdown: func(ctx context.Context) error {
			if err := h.server.Shutdown(ctx); err != nil {
				h.log.Errorf("Server shutdown error: %v", err)
				return err
			}
			h.log.Info("Server gracefully stopped")
			return nil
		},
	}, nil
}
