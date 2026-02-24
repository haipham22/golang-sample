package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	governerrors "github.com/haipham22/govern/errors"
	"go.uber.org/zap"

	"golang-sample/internal/model"
	schemas2 "golang-sample/internal/schemas"
	"golang-sample/internal/storage/user"
	"golang-sample/pkg/utils/password"
)

type impl struct {
	log           *zap.SugaredLogger
	storage       user.Storage
	jwtSecret     string
	jwtExpiration time.Duration
}

func NewAuthService(
	log *zap.SugaredLogger,
	storage user.Storage,
	jwtSecret string,
	jwtExpiration time.Duration,
) Service {
	return &impl{
		log:           log,
		storage:       storage,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

func (s *impl) Register(ctx context.Context, req RegisterRequest) (*model.User, error) {
	usernameExists, emailExists, err := s.storage.CheckUniqueness(ctx, req.Username, req.Email)
	if err != nil {
		s.log.Errorf("Failed to check uniqueness: %v", err)
		return nil, governerrors.WrapCode(governerrors.CodeInternal, err)
	}

	if usernameExists {
		s.log.Warnf("Registration attempted with existing username")
		return nil, governerrors.NewCode(governerrors.CodeConflict, "username already exists")
	}

	if emailExists {
		s.log.Warnf("Registration attempted with existing email")
		return nil, governerrors.NewCode(governerrors.CodeConflict, "email already exists")
	}

	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, governerrors.WrapCode(governerrors.CodeInternal, err)
	}

	m := &model.User{
		Username: req.Username,
		Email:    req.Email,
	}

	createdUser, err := s.storage.CreateUserWithPassword(ctx, m, hashedPassword)
	if err != nil {
		s.log.Errorf("Failed to create user: %v", err)
		return nil, governerrors.WrapCode(governerrors.CodeInternal, err)
	}

	s.log.Infof("User registered successfully: %s", createdUser.Username)
	return createdUser, nil
}

func (s *impl) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	account, passwordHash, err := s.storage.FindUserByUsernameWithPassword(ctx, req.Username)
	if err != nil {
		s.log.Errorf("Failed to find account by username: %v", err)
		return nil, governerrors.WrapCode(governerrors.CodeInternal, err)
	}

	if account == nil {
		s.log.Warnf("Login attempted with non-existent username")
		return nil, governerrors.ErrUnauthorized
	}

	if !password.CheckPasswordHash(req.Password, passwordHash) {
		s.log.Warnf("Login attempted with invalid password")
		return nil, governerrors.ErrUnauthorized
	}

	token, expiresAt, err := s.generateToken(account)
	if err != nil {
		s.log.Errorf("Failed to generate token: %v", err)
		return nil, governerrors.WrapCode(governerrors.CodeInternal, err)
	}

	s.log.Infof("User logged in successfully: %s", account.Username)
	return &LoginResponse{
		Token:     token,
		User:      account,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *impl) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiration)

	claims := schemas2.JwtClaims{
		ID:       strconv.FormatUint(uint64(user.ID), 10),
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
