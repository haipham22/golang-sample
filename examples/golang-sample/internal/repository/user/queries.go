package user

import (
	"context"
	stderrors "errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/orm"
)

// ListUsersParams carries optional pagination for user listing. When Limit is
// zero a sensible default is applied.
type ListUsersParams struct {
	Limit  int
	Offset int
}

// defaultUserListLimit caps ListUsers when no limit is provided.
const defaultUserListLimit = 100

// FindUserByID loads a user by primary key. Returns (nil, nil) when the user
// does not exist, mirroring FindUserByUsername's not-found contract so callers
// can distinguish "absent" from "hard error".
func (s *repo) FindUserByID(ctx context.Context, id uint) (*domain.User, error) {
	var ormUser orm.User
	err := s.db.WithContext(ctx).First(&ormUser, id).Error
	if err != nil && stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		s.log.Errorf("Failed to find user by id %d, err: %#v", id, err)
		return nil, apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return ormToModel(&ormUser), nil
}

// ListUsers returns a paginated slice of users together with the total count.
// Results are ordered by ID ascending for stable pagination.
func (s *repo) ListUsers(ctx context.Context, params ListUsersParams) ([]*domain.User, int64, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = defaultUserListLimit
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&orm.User{}).Count(&total).Error; err != nil {
		s.log.Errorf("Failed to count users, err: %#v", err)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var ormUsers []*orm.User
	query := s.db.WithContext(ctx).
		Model(&orm.User{}).
		Order("id ASC").
		Limit(limit)
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}
	if err := query.Find(&ormUsers).Error; err != nil {
		s.log.Errorf("Failed to list users, err: %#v", err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return ormSliceToModelSlice(ormUsers), total, nil
}
