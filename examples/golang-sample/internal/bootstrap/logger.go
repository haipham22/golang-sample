package bootstrap

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/pkg/config"
)

// NewLogger builds a zap SugaredLogger keyed off the resolved config. The
// returned cleanup syncs the logger and must be deferred by the caller.
//
// Production builds use a production encoder; development/debug builds use the
// colored development encoder. The global zap logger is NOT replaced here —
// cmd/ owns global state.
func NewLogger(cfg *config.EnvConfigMap) (*zap.SugaredLogger, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("logger: nil config")
	}

	var (
		logger *zap.Logger
		err    error
	)
	if cfg.App.Env == config.EnvProduction {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("logger: build: %w", err)
	}

	cleanup := func() { _ = logger.Sync() }
	return logger.Sugar(), cleanup, nil
}
