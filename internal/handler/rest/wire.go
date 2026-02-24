//go:build wireinject
// +build wireinject

package rest

import (
	"time"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	governhttp "github.com/haipham22/govern/http"

	authctrl "golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "golang-sample/internal/handler/rest/controllers/health"
	authservice "golang-sample/internal/service/auth"
	userRepo "golang-sample/internal/storage/user"
	"golang-sample/pkg/config"
	"golang-sample/pkg/postgres"
)

// authConfig holds JWT configuration
type authConfig struct {
	jwtSecret string
}

func provideAuthService(
	log *zap.SugaredLogger,
	storage userRepo.Storage,
	cfg authConfig,
) authservice.Service {
	jwtExpiration := 72 * time.Hour
	return authservice.NewAuthService(log, storage, cfg.jwtSecret, jwtExpiration)
}

func provideDebugFlag(appConfig *config.EnvConfigMap) bool {
	return appConfig.App.Debug
}

func provideEnv(appConfig *config.EnvConfigMap) string {
	return appConfig.App.Env
}

func provideDB(appConfig *config.EnvConfigMap) (*gorm.DB, func(), error) {
	return postgres.NewGormDB(appConfig.Postgres.DSN)
}

// provideAuthConfig extracts JWT config from main config
func provideAuthConfig(appConfig *config.EnvConfigMap) authConfig {
	if appConfig.API.Secret == "" {
		panic("JWT secret is required but not configured. Please set api.secret in your config file.")
	}
	return authConfig{
		jwtSecret: appConfig.API.Secret,
	}
}

// New creates a new Handler with all dependencies wired.
// Returns: server, cleanup function, error
func New(
	log *zap.SugaredLogger,
	port int64,
	appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
	panic(wire.Build(
		// Config providers
		wire.NewSet(provideAuthConfig),

		// Database
		wire.NewSet(provideDB),
		wire.NewSet(userRepo.New),

		// Services
		wire.NewSet(provideAuthService),

		// Controllers
		wire.NewSet(authctrl.New),
		wire.NewSet(healthctrl.New),

		wire.NewSet(provideDebugFlag),
		wire.NewSet(provideEnv),

		// HTTP Server
		wire.NewSet(NewHandler),

		echo.New,
	))
}
