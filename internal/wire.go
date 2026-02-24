//go:build wireinject
// +build wireinject

package internal

import (
	"golang-sample/internal/handler/rest"
	"golang-sample/internal/handler/rest/auth"
	"golang-sample/internal/handler/rest/health"
	"golang-sample/internal/storage"
	"golang-sample/pkg/postgres"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func New(
	dbDSN string,
	log *zap.SugaredLogger,
) (*rest.Handler, func(), error) {
	panic(wire.Build(
		rest.NewHandler,
		echo.New,
		governpostgres.New,
		wire.NewSet(storage.NewStorage),
		wire.NewSet(auth.NewAuthController),
		wire.NewSet(health.NewController),
	))

	return &rest.Handler{}, nil, nil
}
