package storage

import (
	"context"
	"fmt"
	"golang-sample/internal/models"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *storageHandler) IsExistBy(field string, condition string) (bool, error) {
	// Check if the username exists in the database
	var count int64
	query := fmt.Sprintf("%s = ?", field)
	if err := s.db.Model(&models.User{}).Where(query, condition).Count(&count).Error; err != nil {
		s.log.Errorf("Failed to check if %s exists, err: %#v", field, zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (s *storageHandler) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		s.log.Errorf("Failed to create user, err: %#v", zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *storageHandler) FindUserByUsername(ctx context.Context, username string) (user *models.User, err error) {
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return
}
