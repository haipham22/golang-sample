package storage

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Storage struct {
	log *zap.SugaredLogger
	db  *gorm.DB
}

func NewStorage(log *zap.SugaredLogger, db *gorm.DB) *Storage {
	return &Storage{
		log: log,
		db:  db,
	}
}
