package rest

import (
	"context"
	"fmt"
	middlewares2 "golang-sample/internal/handler/rest/middlewares"
	"net/http"
	"time"

	"golang-sample/internal/handler/rest/auth"
	"golang-sample/internal/handler/rest/health"
	apiValidator "golang-sample/internal/validator"

	governerrors "github.com/haipham22/govern/errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"golang-sample/pkg/config"
)

type (
	Handler struct {
		log    *zap.SugaredLogger
		server *echo.Echo

		auth   *auth.Controller
		health *health.Controller
	}

	ServerFunc struct {
		Start    func() error
		Shutdown func(context.Context) error
	}
)

const (
	readHeaderTimeout = 30 * time.Second
)

func NewHandler(log *zap.SugaredLogger, e *echo.Echo, auth *auth.Controller, health *health.Controller) *Handler {

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
		middlewares2.Logger(zap.L()),
		middlewares2.MetricsMiddleware(),
	)

	// Set custom HTTP error handler with govern error code support
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	return &Handler{
		log:    log,
		server: e,
		auth:   auth,
		health: health,
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

// customHTTPErrorHandler handles errors with govern error code support
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	// Check govern error codes first
	if errCode, ok := governerrors.GetCode(err); ok {
		switch errCode {
		case governerrors.CodeInvalid:
			code = http.StatusBadRequest
			message = err.Error()
		case governerrors.CodeNotFound:
			code = http.StatusNotFound
			message = "Resource not found"
		case governerrors.CodeUnauthorized:
			code = http.StatusUnauthorized
			message = "Unauthorized"
		case governerrors.CodeForbidden:
			code = http.StatusForbidden
			message = "Forbidden"
		case governerrors.CodeConflict:
			code = http.StatusConflict
			message = err.Error()
		default:
			// Log unknown error codes
			c.Logger().Error("Unknown error code", zap.String("code", string(errCode)), zap.Error(err))
		}
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if he.Message != nil {
			message = fmt.Sprintf("%v", he.Message)
		}
	} else if err != nil {
		message = err.Error()
	}

	// Log error
	c.Logger().Error("Request error",
		zap.String("path", c.Path()),
		zap.Int("status", code),
		zap.Error(err),
	)

	// Send response
	if !c.Response().Committed {
		if code >= 500 {
			c.JSON(code, map[string]interface{}{
				"error": "Internal Server Error",
				"path":  c.Path(),
			})
		} else {
			c.JSON(code, map[string]interface{}{
				"error": message,
				"path":  c.Path(),
			})
		}
	}
}
