package user

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-sample/internal/model"
	"golang-sample/internal/orm"
)

func (s *repo) IsExistBy(ctx context.Context, field string, condition string) (bool, error) {
	// Whitelist of allowed columns to prevent SQL injection
	allowedColumns := map[string]bool{
		"username": true,
		"email":    true,
		"id":       true,
	}

	if !allowedColumns[field] {
		s.log.Errorf("Invalid field name for existence check: %s", field)
		return false, fmt.Errorf("invalid field name: %s", field)
	}

	// Check if the field exists in the database
	var count int64
	query := fmt.Sprintf("%s = ?", field)
	if err := s.db.WithContext(ctx).Model(&orm.User{}).Where(query, condition).Count(&count).Error; err != nil {
		s.log.Errorf("Failed to check if %s exists, err: %#v", field, zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// CheckUniqueness checks both username and email uniqueness in a single optimized query.
// Uses CASE WHEN conditional aggregation to check both fields in one database roundtrip.
// Returns (usernameExists, emailExists, error)
func (s *repo) CheckUniqueness(ctx context.Context, username, email string) (bool, bool, error) {
	type UniquenessResult struct {
		UsernameCount int64
		EmailCount    int64
	}

	var result UniquenessResult
	err := s.db.WithContext(ctx).Model(&orm.User{}).Select(`
		COUNT(CASE WHEN username = ? THEN 1 END) as username_count,
		COUNT(CASE WHEN email = ? THEN 1 END) as email_count
	`, username, email).Scan(&result).Error

	if err != nil {
		s.log.Errorf("Failed to check uniqueness, err: %#v", zap.Error(err))
		return false, false, err
	}

	return result.UsernameCount > 0, result.EmailCount > 0, nil
}

func (s *repo) CreateUserWithPassword(ctx context.Context, user *model.User, passwordHash string) (*model.User, error) {
	// Convert domain model to ORM
	ormUser := modelToORM(user)
	ormUser.PasswordHash = passwordHash

	if err := s.db.WithContext(ctx).Create(&ormUser).Error; err != nil {
		s.log.Errorf("Failed to create user, err: %#v", zap.Error(err))
		return nil, err
	}

	// Convert back to domain model (without password)
	return ormToModel(ormUser), nil
}

func (s *repo) FindUserByUsername(ctx context.Context, username string) (user *model.User, err error) {
	var ormUser *orm.User
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&ormUser).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert ORM to domain model
	return ormToModel(ormUser), nil
}

func (s *repo) FindUserByUsernameWithPassword(ctx context.Context, username string) (user *model.User, passwordHash string, err error) {
	var ormUser *orm.User
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&ormUser).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	// Convert ORM to domain model
	return ormToModel(ormUser), ormUser.PasswordHash, nil
}
