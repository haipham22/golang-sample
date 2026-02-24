package middlewares

import (
	"fmt"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ignoresURL = []string{
	"/health",
}

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			return slices.Contains(ignoresURL, c.Request().URL.Path)
		},
		HandleError:     true,
		LogURI:          true,
		LogStatus:       true,
		LogRemoteIP:     true,
		LogLatency:      true,
		LogHost:         true,
		LogMethod:       true,
		LogResponseSize: true,
		LogUserAgent:    true,
		LogRequestID:    true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {

			fields := []zapcore.Field{
				zap.String("remote_ip", v.RemoteIP),
				zap.Duration("latency", v.Latency),
				zap.String("host", v.Host),
				zap.String("request", fmt.Sprintf("%s %s", v.Method, v.URI)),
				zap.Int("status", v.Status),
				zap.Int64("size", v.ResponseSize),
				zap.String("user_agent", v.UserAgent),
				zap.String("request_id", v.RequestID),
			}

			n := v.Status
			switch {
			case n >= 500:
				logger.With(zap.Error(v.Error)).Error("Server error", fields...)
			case n >= 400:
				logger.With(zap.Error(v.Error)).Warn("Client error", fields...)
			case n >= 300:
				logger.Info("Redirection", fields...)
			default:
				logger.Info("Success", fields...)
			}

			return nil
		},
	})
}
