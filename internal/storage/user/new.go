package user

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-sample/internal/model"
)

type Storage interface {
	IsExistBy(ctx context.Context, field string, condition string) (bool, error)
	// CheckUniqueness checks both username and email uniqueness in a single query
	// Returns (usernameExists, emailExists, error)
	CheckUniqueness(ctx context.Context, username, email string) (bool, bool, error)
	// CreateUserWithPassword creates a user with password hash (returns domain model without password)
	CreateUserWithPassword(ctx context.Context, user *model.User, passwordHash string) (*model.User, error)
	FindUserByUsername(ctx context.Context, username string) (user *model.User, err error)
	// FindUserByUsernameWithPassword finds user and returns with password hash for authentication
	FindUserByUsernameWithPassword(ctx context.Context, username string) (user *model.User, passwordHash string, err error)
}

type repo struct {
	log *zap.SugaredLogger
	db  *gorm.DB
}

func New(log *zap.SugaredLogger, db *gorm.DB) Storage {
	return &repo{
		log: log,
		db:  db,
	}
}
