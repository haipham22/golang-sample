package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// impl is the default Service implementation.
type impl struct {
	log  *zap.SugaredLogger
	repo Repository
}

// Compile-time guard.
var _ Service = (*impl)(nil)

// NewService wires a user Service with its repository dependency.
func NewService(log *zap.SugaredLogger, repo Repository) Service {
	return &impl{log: log, repo: repo}
}

// GetByID returns a single user by ID. A zero id is rejected as invalid input;
// a missing user yields apperrors.NotFound.
func (s *impl) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	if id == 0 {
		return nil, apperrors.NewCode(apperrors.CodeInvalid, "user id is required")
	}
	u, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperrors.NotFound("user")
	}
	return u, nil
}

// List returns a page of users as DTOs.
func (s *impl) List(ctx context.Context, params ListParams) (*ListResponse, error) {
	items, total, err := s.repo.ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	dtos := make([]*UserDTO, 0, len(items))
	for _, u := range items {
		dtos = append(dtos, toDTO(u))
	}
	return &ListResponse{Items: dtos, Total: total}, nil
}

// toDTO converts a domain user to its DTO.
func toDTO(u *domain.User) *UserDTO {
	if u == nil {
		return nil
	}
	return &UserDTO{ID: u.ID, Username: u.Username, Email: u.Email}
}
