//go:build wireinject
// +build wireinject

package internal

import (
	"golang-sample/internal/handler/rest/auth"
	"golang-sample/internal/handler/rest/health"
	"golang-sample/internal/storage"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"golang-sample/pkg/postgres"
)

func New(
	dbDSN string,
	log *zap.SugaredLogger,
) (*Handler, func(), error) {
	panic(wire.Build(
		NewHandler,
		echo.New,
		postgres.NewGormDB,
		wire.NewSet(storage.NewStorage),
		wire.NewSet(auth.NewAuthController),
		wire.NewSet(health.NewController),
	))

	return &Handler{}, nil, nil
}
