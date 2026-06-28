// Package bootstrap is the composition root: it wires the logger, config,
// repositories, use cases, and delivery layer and returns the assembled HTTP
// server plus a cleanup func. Manual DI (no code generation) — mirroring the
// former internal/handler/rest/di.go graph but extracted into its own package
// so cmd/ stays thin.
//
// The HTTP wiring still lives in internal/handler/rest (New/NewHandler); this
// package orchestrates it.
package bootstrap

import (
	"fmt"

	governhttp "github.com/haipham22/govern/http"
	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/pkg/config"
)

// Config carries the inputs bootstrap needs to assemble the server.
type Config struct {
	// Port is the HTTP listen port.
	Port int64
	// AppConfig is the resolved environment configuration.
	AppConfig *config.EnvConfigMap
}

// httpServerFactory builds the HTTP server from a logger, port, and config.
// It is a package-level indirection so tests can exercise New() end-to-end
// without a live Postgres (the default wiring, NewHTTPServer, requires one).
// Production code leaves this at its zero value, which delegates to
// NewHTTPServer — the former hard-coded call.
//
// ponytail: a sync.Once-based override would also work; the indirection is the
// simplest seam that preserves the public New signature.
var httpServerFactory = func(log *zap.SugaredLogger, port int64, cfg *config.EnvConfigMap) (governhttp.Server, func(), error) {
	return NewHTTPServer(log, port, cfg)
}

// New assembles the HTTP server: it builds the logger and delegates HTTP/DB/
// service wiring to httpServerFactory (NewHTTPServer by default). The returned
// cleanup closes the DB and syncs the logger; it must be called on shutdown
// (typically via defer in cmd).
//
// The server is returned directly (no App wrapper) because cmd consumes only
// the server — logging for graceful.Run uses the global zap.S() logger, so a
// dedicated App.Log field would be dead weight.
func New(cfg Config) (governhttp.Server, func(), error) {
	if cfg.AppConfig == nil {
		return nil, nil, fmt.Errorf("bootstrap: nil app config")
	}

	log, logCleanup, err := NewLogger(cfg.AppConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap: %w", err)
	}

	server, httpCleanup, err := httpServerFactory(log, cfg.Port, cfg.AppConfig)
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

	return server, cleanup, nil
}
