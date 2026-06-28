package rest

import (
	stderrors "errors"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	governhttp "github.com/haipham22/govern/http"

	authctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/auth"
	healthctrl "github.com/haipham22/golang-sample/internal/handler/rest/controllers/health"
	authservice "github.com/haipham22/golang-sample/internal/usecase/auth"
	userRepo "github.com/haipham22/golang-sample/internal/repository/user"
	"github.com/haipham22/golang-sample/pkg/config"
	"github.com/haipham22/golang-sample/pkg/postgres"
)

// jwtExpiration is how long issued JWT tokens remain valid.
const jwtExpiration = 72 * time.Hour

// ErrMissingJWTSecret is returned when the API secret is not configured.
var ErrMissingJWTSecret = stderrors.New("JWT secret is required but not configured (set api.secret)")

// New creates the HTTP server with all dependencies wired manually (replaces
// the former code-generated Wire injector). It mirrors the former
// wire_gen.go dependency graph:
//
//	appConfig -> db (postgres) -> storage -> auth service -> auth controller
//	appConfig -> db -> health controller
//	appConfig -> debug/env, echo -> NewHandler -> server
//
// Returns the server, a cleanup function that closes the DB, and any error.
func New(
	log *zap.SugaredLogger,
	port int64,
	appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
	// 1. Validate required config.
	if appConfig.API.Secret == "" {
		return nil, nil, ErrMissingJWTSecret
	}

	// 2. Initialize database (returns cleanup on success).
	db, cleanup, err := postgres.NewGormDB(appConfig.Postgres.DSN)
	if err != nil {
		return nil, nil, err
	}

	// 3. Storage layer.
	storage := userRepo.New(log, db)

	// 4. Service layer.
	authService := authservice.NewAuthService(log, storage, appConfig.API.Secret, jwtExpiration)

	// 5. Controllers.
	authController := authctrl.New(authService)
	healthController := healthctrl.New(db)

	// 6. Echo instance + HTTP server.
	e := echo.New()
	server := NewHandler(
		log,
		e,
		authController,
		healthController,
		port,
		appConfig.App.Debug,
		appConfig.App.Env,
	)

	return server, cleanup, nil
}
