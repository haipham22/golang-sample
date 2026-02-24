package storage

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-sample/pkg/models"
)

type Storage interface {
	IsExistBy(field string, condition string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	FindUserByUsername(ctx context.Context, username string) (user *models.User, err error)
}

type storageHandler struct {
	log *zap.SugaredLogger
	db  *gorm.DB
}

func NewStorage(log *zap.SugaredLogger, db *gorm.DB) Storage {
	return &storageHandler{
		log: log,
		db:  db,
	}
}
