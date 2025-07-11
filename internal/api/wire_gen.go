// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang-sample/internal/api/routes/auth"
	"golang-sample/internal/api/storage"
	"golang-sample/pkg/postgres"
)

// Injectors from wire.go:

func New(dbDSN string, log *zap.SugaredLogger) (*Handler, func(), error) {
	echoEcho := echo.New()
	db, cleanup, err := postgres.NewGormDB(dbDSN)
	if err != nil {
		return nil, nil, err
	}
	storageStorage := storage.NewStorage(log, db)
	controller := auth.NewAuthController(log, storageStorage)
	handler := NewHandler(log, echoEcho, controller)
	return handler, func() {
		cleanup()
	}, nil
}
