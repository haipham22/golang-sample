package rest

import (
	"errors"
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

// resolveError maps an error to an HTTP status and a standard response body.
// Order: apperrors-typed errors use centralized resolution (with validation
// detail enrichment), echo.HTTPError values are sanitized, everything else is
// a generic 500.
func resolveError(err error, path, requestID string) (int, apperrors.Response) {
	// 1. apperrors-typed error: centralized status + body mapping.
	if code, ok := apperrors.GetCode(err); ok {
		status, body := apperrors.Resolve(err, path, requestID)
		if code == apperrors.CodeInvalid {
			enrichValidation(&body, err)
		}
		return status, body
	}

	// 2. Echo HTTP error: pass through status, sanitize 5xx messages.
	if he, ok := err.(*echo.HTTPError); ok {
		clientMsg := he.Message
		if he.Code >= 500 {
			clientMsg = http.StatusText(http.StatusInternalServerError)
		}
		if clientMsg == "" {
			clientMsg = http.StatusText(he.Code)
		}
		return he.Code, apperrors.Response{
			Msg:       clientMsg,
			Error:     clientMsg,
			Path:      path,
			RequestID: requestID,
		}
	}

	// 2b. Echo sentinel errors (ErrNotFound, ErrBadRequest, …) are an unexported
	// *httpError type that implements HTTPStatusCoder but is not *HTTPError, so
	// the branch above misses them. echo.StatusCode handles both via errors.As.
	if code := echo.StatusCode(err); code != 0 {
		msg := http.StatusText(code)
		if code >= 500 {
			msg = http.StatusText(http.StatusInternalServerError)
		}
		return code, apperrors.Response{
			Msg:       msg,
			Error:     msg,
			Path:      path,
			RequestID: requestID,
		}
	}

	// 3. Unknown error: generic internal server error.
	return http.StatusInternalServerError, apperrors.Response{
		Msg:       http.StatusText(http.StatusInternalServerError),
		Error:     http.StatusText(http.StatusInternalServerError),
		Path:      path,
		RequestID: requestID,
	}
}

// enrichValidation fills in field-level details when err wraps a
// validator.ValidationError; otherwise leaves the generic invalid-input body.
func enrichValidation(body *apperrors.Response, err error) {
	var validationErr *apiValidator.ValidationError
	if errors.As(err, &validationErr) {
		body.Msg = validationErr.Detail.Msg
		body.Error = validationErr.Detail.Msg
		body.Errors = []apperrors.FieldError{{
			Property: validationErr.Detail.Property,
			Msg:      validationErr.Detail.Msg,
		}}
	}
}
