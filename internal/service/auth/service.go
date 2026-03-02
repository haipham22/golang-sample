package auth

import (
	"context"
	"time"

	"golang-sample/internal/model"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*model.User, error)
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
	User      *model.User
	ExpiresAt time.Time
}
