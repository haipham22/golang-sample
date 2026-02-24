package rest

import (
	"errors"
	"fmt"
	"golang-sample/pkg/config"
	"net/http"
	"time"

	governerrors "github.com/haipham22/govern/errors"
	governhttp "github.com/haipham22/govern/http"
	httpEcho "github.com/haipham22/govern/http/echo"
	"github.com/haipham22/govern/http/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	authctrl "golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "golang-sample/internal/handler/rest/controllers/health"
	"golang-sample/internal/handler/rest/middlewares"
	apiValidator "golang-sample/internal/validator"
)

func NewHandler(
	log *zap.SugaredLogger,
	e *echo.Echo,
	authCtrl *authctrl.Controller,
	healthCtrl *healthctrl.Controller,
	port int64,
	debug bool,
	env string,
) governhttp.Server {

	e.Validator = apiValidator.NewCustomValidator()

	if debug {
		e.Debug = true
	}

	e.Use(
		echomiddleware.RemoveTrailingSlashWithConfig(echomiddleware.TrailingSlashConfig{
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

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	// Create an HTTP server
	e = initRouter(e, authCtrl, healthCtrl)

	server := governhttp.NewServer(
		fmt.Sprintf(":%d", port),
		e,
		governhttp.WithTimeout(30*time.Second, 60*time.Second, 120*time.Second),
		governhttp.WithLogger(log),
	)

	return server
}

// customHTTPErrorHandler handles errors with govern error code support
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	var responseBody interface{}

	// Check govern error codes first
	if errCode, ok := governerrors.GetCode(err); ok {
		switch errCode {
		case governerrors.CodeInvalid:
			code = http.StatusBadRequest
			// Try to extract validation error details
			responseBody = buildValidationErrorResponse(err, c.Path())
		case governerrors.CodeNotFound:
			code = http.StatusNotFound
			responseBody = map[string]interface{}{
				"msg":   "Resource not found",
				"error": "Resource not found",
				"path":  c.Path(),
			}
		case governerrors.CodeUnauthorized:
			code = http.StatusUnauthorized
			responseBody = map[string]interface{}{
				"msg":   "Unauthorized",
				"error": "Unauthorized",
				"path":  c.Path(),
			}
		case governerrors.CodeForbidden:
			code = http.StatusForbidden
			responseBody = map[string]interface{}{
				"msg":   "Forbidden",
				"error": "Forbidden",
				"path":  c.Path(),
			}
		case governerrors.CodeConflict:
			code = http.StatusConflict
			responseBody = map[string]interface{}{
				"msg":   "Resource already exists",
				"error": "conflict occurred",
				"path":  c.Path(),
			}
		case governerrors.CodeInternal:
			code = http.StatusInternalServerError
			responseBody = map[string]interface{}{
				"msg":   "Internal Server Error",
				"error": "Internal Server Error",
				"path":  c.Path(),
			}
		default:
			// Log unknown error codes
			zap.L().Error("Unknown error code in error handler",
				zap.String("code", string(errCode)),
				zap.String("path", c.Path()),
				zap.Error(err))
			code = http.StatusInternalServerError
			responseBody = map[string]interface{}{
				"msg":   "Internal Server Error",
				"error": "Internal Server Error",
				"path":  c.Path(),
			}
		}
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code

		// Sanitize 5xx error messages to avoid leaking internal details
		var clientMsg string
		if code >= 500 {
			clientMsg = "Internal Server Error"
			// Log the actual internal error message
			zap.L().Error("HTTPError (5xx)",
				zap.Int("status", code),
				zap.String("path", c.Path()),
				zap.String("internal_message", fmt.Sprintf("%v", he.Message)),
			)
		} else {
			clientMsg = fmt.Sprintf("%v", he.Message)
		}

		responseBody = map[string]interface{}{
			"msg":   clientMsg,
			"error": clientMsg,
			"path":  c.Path(),
		}
	} else if err != nil {
		code = http.StatusInternalServerError
		responseBody = map[string]interface{}{
			"msg":   "Internal Server Error",
			"error": "Internal Server Error",
			"path":  c.Path(),
		}
	}

	// Log error
	zap.L().Error("Request error",
		zap.String("path", c.Path()),
		zap.Int("status", code),
		zap.Error(err),
	)

	// Send response
	if !c.Response().Committed {
		c.JSON(code, responseBody)
	}
}

// buildValidationErrorResponse builds a detailed validation error response
func buildValidationErrorResponse(err error, path string) map[string]interface{} {
	// Try to unwrap and find ValidationError
	var validationErr *apiValidator.ValidationError
	if errors.As(err, &validationErr) {
		return map[string]interface{}{
			"msg":   validationErr.Detail.Msg,
			"error": validationErr.Detail.Msg,
			"errors": []map[string]interface{}{
				{
					"property": validationErr.Detail.Property,
					"msg":      validationErr.Detail.Msg,
				},
			},
			"path": path,
		}
	}

	// Fallback to generic error message for security
	return map[string]interface{}{
		"msg":   "invalid request parameters",
		"error": "invalid request parameters",
		"path":  path,
	}
}
