// Package user implements read-only user use cases (lookup + listing). The
// Repository interface is defined here (consumer-defined, bxcodec/clean-arch
// pattern); the existing internal/repository/user package satisfies it via the
// FindUserByID/ListUsers methods.
package user

import (
	"context"

	"github.com/haipham22/golang-sample/internal/domain"
)

// Repository is the user use case's persistence port. Defined by the consumer
// (this package), not by the domain or repository layer.
type Repository interface {
	// FindUserByID returns the user with the given ID, or (nil, nil) when absent.
	FindUserByID(ctx context.Context, id uint) (*domain.User, error)
	// ListUsers returns a page of users plus the total count.
	ListUsers(ctx context.Context, params ListParams) ([]*domain.User, int64, error)
}

// ListParams carries pagination for user listing.
type ListParams struct {
	Limit  int
	Offset int
}

// Service is the user use case contract.
type Service interface {
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	List(ctx context.Context, params ListParams) (*ListResponse, error)
}

// UserDTO is the API-facing user representation returned by use cases.
type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ListResponse wraps a page of users with the total count.
type ListResponse struct {
	Items []*UserDTO
	Total int64
}
