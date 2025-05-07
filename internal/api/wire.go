//go:build wireinject
// +build wireinject

package api

import (
	"github.com/google/wire"
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
		postgres.NewGormDB,
		wire.NewSet(storage.NewStorage),
		wire.NewSet(auth.NewAuthController),
	))

	return &Handler{}, nil, nil
}
