package auth

import (
	"context"
	"time"

	"github.com/haipham22/golang-sample/internal/domain"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
	FullName string
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Token     string
	User      *domain.User
	ExpiresAt time.Time
}
