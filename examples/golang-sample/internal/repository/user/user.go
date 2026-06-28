package user

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/orm"
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
		return false, apperrors.InvalidInput(fmt.Sprintf("invalid field name: %s", field))
	}

	// Check if the field exists in the database
	var count int64
	query := fmt.Sprintf("%s = ?", field)
	if err := s.db.WithContext(ctx).Model(&orm.User{}).Where(query, condition).Count(&count).Error; err != nil {
		s.log.Errorf("Failed to check if %s exists, err: %#v", field, zap.Error(err))
		return false, apperrors.WrapCode(apperrors.CodeInternal, err)
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
		return false, false, apperrors.WrapCode(apperrors.CodeInternal, err)
	}

	return result.UsernameCount > 0, result.EmailCount > 0, nil
}

func (s *repo) CreateUserWithPassword(ctx context.Context, user *domain.User, passwordHash string) (*domain.User, error) {
	// Convert domain model to ORM
	ormUser := modelToORM(user)
	ormUser.PasswordHash = passwordHash

	if err := s.db.WithContext(ctx).Create(&ormUser).Error; err != nil {
		// Wrap as Internal but preserve the underlying error chain so the
		// service layer can still detect duplicate-key race conditions via
		// errors.Is(err, gorm.ErrDuplicatedKey) / message inspection.
		s.log.Errorf("Failed to create user, err: %#v", zap.Error(err))
		return nil, apperrors.WrapCode(apperrors.CodeInternal, err)
	}

	// Convert back to domain model (without password)
	return ormToModel(ormUser), nil
}

func (s *repo) FindUserByUsername(ctx context.Context, username string) (user *domain.User, err error) {
	var ormUser *orm.User
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&ormUser).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Not-found is surfaced as (nil, nil); the caller distinguishes via the
		// returned user pointer rather than an error code.
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.WrapCode(apperrors.CodeInternal, err)
	}

	// Convert ORM to domain model
	return ormToModel(ormUser), nil
}

func (s *repo) FindUserByUsernameWithPassword(ctx context.Context, username string) (user *domain.User, passwordHash string, err error) {
	var ormUser *orm.User
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&ormUser).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", apperrors.WrapCode(apperrors.CodeInternal, err)
	}

	// Convert ORM to domain model
	return ormToModel(ormUser), ormUser.PasswordHash, nil
}
