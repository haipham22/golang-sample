package auth

import (
	"go.uber.org/zap"

	"golang-sample/internal/api/storage"
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
