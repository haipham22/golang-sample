package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	"golang-sample/internal/api/middlewares"
	"golang-sample/internal/api/routes/auth"
	apiValidator "golang-sample/internal/api/validator"
	"golang-sample/pkg/config"
)

type Handler struct {
	log  *zap.SugaredLogger
	auth *auth.Controller
}

func NewApiBiz(log *zap.SugaredLogger, auth *auth.Controller) *Handler {
	return &Handler{
		log:  log,
		auth: auth,
	}
}

func (h *Handler) ServeHTTP() *echo.Echo {
	e := echo.New()

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

	public := e.Group("/api")

	SetAuthRoutes(public, h.auth)

	return e
}
