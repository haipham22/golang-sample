package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/haipham22/golang-sample/pkg/config"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	governhttp "github.com/haipham22/govern/http"
	httpEcho "github.com/haipham22/govern/http/echo"
	"github.com/haipham22/govern/http/middleware"

	authctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
	productctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/product"
	"github.com/haipham22/golang-sample/internal/handler/rest/middlewares"
	apiValidator "github.com/haipham22/golang-sample/internal/validator"
)

func NewHandler(
	log *zap.SugaredLogger,
	e *echo.Echo,
	authCtrl *authctrl.Controller,
	healthCtrl *healthctrl.Controller,
	productCtrl *productctrl.Controller,
	port int64,
	debug bool,
	env string,
) governhttp.Server {

	e.Validator = apiValidator.NewCustomValidator()

	// Echo v5 removed the Echo.Debug toggle; debug behavior is now controlled
	// per-middleware (e.g. logger level, error-handler detail). The `debug`
	// flag still gates Swagger below.

	e.Use(
		echomiddleware.RemoveTrailingSlashWithConfig(echomiddleware.RemoveTrailingSlashConfig{
			RedirectCode: http.StatusPermanentRedirect,
		}),
		echomiddleware.Recover(),
		echomiddleware.RequestID(),
		middlewares.BodyLimit(),
		middleware.TrimStrings,
		middlewares.SecurityHeaders(),
		middlewares.CORS(),
	)

	httpEcho.WithEchoSwagger(
		e,
		httpEcho.WithSwaggerEnabled(debug && env != config.EnvProduction),
		httpEcho.WithSwaggerPath("/docs/*"),
	)

	e.HTTPErrorHandler = makeHTTPErrorHandler(log)

	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	// Create an HTTP server
	e = initRouter(e, authCtrl, healthCtrl, productCtrl)

	server := governhttp.NewServer(
		fmt.Sprintf(":%d", port),
		e,
		governhttp.WithTimeout(30*time.Second, 60*time.Second, 120*time.Second),
		governhttp.WithLogger(log),
	)

	return server
}

// makeHTTPErrorHandler returns an Echo HTTP error handler that maps errors to a
// standardized JSON response (apperrors.Response) with sanitized client
// messages, request-ID propagation, and structured logging via the injected
// logger (replacing the previous global-zap switch handler).
func makeHTTPErrorHandler(log *zap.SugaredLogger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		path := c.Path()
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)

		status, body := resolveError(err, path, requestID)

		apperrors.LogRequestError(log, err, path, status)

		// Echo v5: c.Response() returns a plain http.ResponseWriter; unwrap to
		// the *echo.Response to read Committed.
		if r, _ := echo.UnwrapResponse(c.Response()); r == nil || !r.Committed {
			c.JSON(status, body)
		}
	}
}

// resolveError maps delivery-specific Echo errors first, then delegates typed
// application errors and unknown fallback to apperrors.Resolve.
func resolveError(err error, path, requestID string) (int, apperrors.Response) {
	if he, ok := err.(*echo.HTTPError); ok {
		return resolveEchoError(he.Code, he.Message, path, requestID)
	}

	// Echo sentinel errors (ErrNotFound, ErrBadRequest, …) are unexported types
	// that implement HTTPStatusCoder. echo.StatusCode handles them via errors.As.
	if code := echo.StatusCode(err); code != 0 {
		return resolveEchoError(code, http.StatusText(code), path, requestID)
	}

	return apperrors.Resolve(err, path, requestID)
}

func resolveEchoError(status int, message string, path, requestID string) (int, apperrors.Response) {
	clientMsg := message
	if status >= 500 {
		clientMsg = http.StatusText(http.StatusInternalServerError)
	}
	if clientMsg == "" {
		clientMsg = http.StatusText(status)
	}
	return status, apperrors.NewBody(clientMsg, path, requestID)
}
