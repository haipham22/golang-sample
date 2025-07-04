//go:build wireinject
// +build wireinject

package api

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"golang-sample/internal/api/routes/auth"
	"golang-sample/internal/api/storage"
	"golang-sample/pkg/postgres"
)

func InitApp(
	isDebugMode bool,
	db string,
	log *zap.SugaredLogger,
) (*Handler, func(), error) {
	panic(wire.Build(
		NewApiBiz,
		echo.New,
		postgres.NewGormDB,
		wire.NewSet(storage.NewStorage),
		wire.NewSet(auth.NewAuthController),
	))

	return &Handler{}, nil, nil
}
