// Package bootstrap is the composition root: it wires the logger, config,
// repositories, use cases, and delivery layer into an App. Manual DI (no code
// generation) — mirroring the former internal/handler/rest/di.go graph but
// extracted into its own package so cmd/ stays thin.
//
// The HTTP wiring still lives in internal/handler/rest (New/NewHandler); this
// package orchestrates it and returns a single App value plus a cleanup func.
package bootstrap

import (
	"fmt"

	"go.uber.org/zap"

	governhttp "github.com/haipham22/govern/http"

	"github.com/haipham22/golang-sample/pkg/config"
)

// App is the assembled application. The HTTP Server is the primary service;
// Log is exposed so cmd/ can use it for graceful.Run.
type App struct {
	HTTPServer governhttp.Server
	Log        *zap.SugaredLogger
}

// Config carries the inputs bootstrap needs to assemble the App.
type Config struct {
	// Port is the HTTP listen port.
	Port int64
	// AppConfig is the resolved environment configuration.
	AppConfig *config.EnvConfigMap
}

// New assembles the App: it builds the logger and delegates HTTP/DB/service
// wiring to internal/handler/rest.New. The returned cleanup closes the DB and
// must be called on shutdown (typically via defer).
func New(cfg Config) (*App, func(), error) {
	if cfg.AppConfig == nil {
		return nil, nil, fmt.Errorf("bootstrap: nil app config")
	}

	log, logCleanup, err := NewLogger(cfg.AppConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap: %w", err)
	}

	server, httpCleanup, err := NewHTTPServer(log, cfg.Port, cfg.AppConfig)
	if err != nil {
		logCleanup()
		return nil, nil, fmt.Errorf("bootstrap: %w", err)
	}

	cleanup := func() {
		if httpCleanup != nil {
			httpCleanup()
		}
		// logCleanup syncs the underlying zap core; safe to call after HTTP
		// tear-down.
		logCleanup()
	}

	return &App{HTTPServer: server, Log: log}, cleanup, nil
}
