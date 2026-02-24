package auth

import (
	"golang-sample/internal/storage"

	"go.uber.org/zap"
)

type Controller struct {
	log     *zap.SugaredLogger
	storage storage.Storage
}

func NewAuthController(log *zap.SugaredLogger, storage storage.Storage) *Controller {
	return &Controller{
		log:     log,
		storage: storage,
	}
}
