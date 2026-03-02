package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	governerrors "github.com/haipham22/govern/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	storageMocks "golang-sample/internal/mocks/storage"
	"golang-sample/internal/model"
	"golang-sample/pkg/utils/password"
)

// Test isolation: Use short JWT expiration for tests (1 hour instead of 72)
// Following Uber: "Use constants for test configuration"
const testJWTExpiration = 1 * time.Hour

// newTestService creates a test service with mocked storage
// Following Uber: "Prefer test helpers over setup duplication"
func newTestService(t *testing.T, storage *storageMocks.MockStorage) Service {
	t.Helper() // Mark as test helper for better stack traces
	log := zap.NewNop().Sugar()
	return NewAuthService(log, storage, "test-secret", testJWTExpiration)
}

// newMockUser creates a test user with hashed password
// Following Uber: "Use builder patterns for test data"
// Returns (model.User, passwordHash) for testing authentication
func newMockUser(t *testing.T, username, plainPassword string) (*model.User, string) {
	t.Helper()
	hash, err := password.HashPassword(plainPassword)
	require.NoError(t, err)
	return &model.User{
		ID:       1,
		Username: username,
		Email:    username + "@example.com",
	}, hash
}

// Table-driven test for Register validation errors
// Following Uber: "Use table-driven tests for multiple cases"
func TestService_Register_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		fullName    string
		setupMock   func(*storageMocks.MockStorage)
		wantErrCode governerrors.ErrorCode
		wantErrMsg  string
	}{
		{
			name:     "username already exists",
			username: "existinguser",
			email:    "new@example.com",
			password: "password",
			fullName: "Test User",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().CheckUniqueness(mock.Anything, "existinguser", "new@example.com").Return(true, false, nil)
			},
			wantErrCode: governerrors.CodeConflict,
			wantErrMsg:  "username",
		},
		{
			name:     "email already exists",
			username: "newuser",
			email:    "existing@example.com",
			password: "password",
			fullName: "Test User",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().CheckUniqueness(mock.Anything, "newuser", "existing@example.com").Return(false, true, nil)
			},
			wantErrCode: governerrors.CodeConflict,
			wantErrMsg:  "email",
		},
		{
			name:     "storage error on uniqueness check",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			fullName: "Test User",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().CheckUniqueness(mock.Anything, "testuser", "test@example.com").Return(false, false, assert.AnError)
			},
			wantErrCode: governerrors.CodeInternal,
			wantErrMsg:  "",
		},
		{
			name:     "storage error on create user",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			fullName: "Test User",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().CheckUniqueness(mock.Anything, "testuser", "test@example.com").Return(false, false, nil)
				m.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*model.User"), mock.AnythingOfType("string")).Return(nil, assert.AnError)
			},
			wantErrCode: governerrors.CodeInternal,
			wantErrMsg:  "",
		},
	}

	for _, tt := range tests {
		tt := tt // Following Uber: "capture range variable"
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Following Uber: "use t.Parallel for independent tests"

			mockStorage := storageMocks.NewMockStorage(t)
			tt.setupMock(mockStorage)

			service := newTestService(t, mockStorage)

			req := RegisterRequest{
				Username: tt.username,
				Email:    tt.email,
				Password: tt.password,
				FullName: tt.fullName,
			}

			user, err := service.Register(context.Background(), req)

			assert.Error(t, err)
			assert.Nil(t, user)

			if tt.wantErrCode != governerrors.CodeInternal {
				var govErr *governerrors.ErrorWithCode
				assert.ErrorAs(t, err, &govErr)
				assert.Equal(t, tt.wantErrCode, govErr.Code)
				if tt.wantErrMsg != "" {
					assert.Contains(t, govErr.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

func TestService_Register_Success(t *testing.T) {
	t.Run("successfully registers a new user", func(t *testing.T) {
		t.Parallel()

		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().CheckUniqueness(mock.Anything, "testuser", "test@example.com").Return(false, false, nil)
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*model.User"), mock.AnythingOfType("string")).RunAndReturn(func(ctx context.Context, user *model.User, passwordHash string) (*model.User, error) {
			// Following Uber: "Verify important invariants in mocks"
			assert.NotEmpty(t, passwordHash, "password should be hashed")
			assert.NotEqual(t, "SecurePass123!", passwordHash, "password hash should not equal plaintext")
			user.ID = 1
			return user, nil
		})

		service := newTestService(t, mockStorage)

		req := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "SecurePass123!",
			FullName: "Test User",
		}

		gotUser, err := service.Register(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, gotUser)
		assert.Equal(t, uint(1), gotUser.ID)
		assert.Equal(t, "testuser", gotUser.Username)
		assert.Equal(t, "test@example.com", gotUser.Email)
		// model.User doesn't have PasswordHash field - security by design
	})
}

// Table-driven test for Login errors
// Following Uber: "Group related error cases"
func TestService_Login_Errors(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		setupMock  func(*storageMocks.MockStorage)
		wantErr    error
		checkToken bool
	}{
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "nonexistent").Return(nil, "", nil)
			},
			wantErr: governerrors.ErrUnauthorized,
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpass",
			setupMock: func(m *storageMocks.MockStorage) {
				mockUser, passwordHash := newMockUser(t, "testuser", "correctpass")
				m.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "testuser").Return(mockUser, passwordHash, nil)
			},
			wantErr: governerrors.ErrUnauthorized,
		},
		{
			name:     "storage error",
			username: "testuser",
			password: "password",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "testuser").Return(nil, "", assert.AnError)
			},
			wantErr: nil, // Service wraps storage errors in ErrorWithCode
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := storageMocks.NewMockStorage(t)
			tt.setupMock(mockStorage)

			service := newTestService(t, mockStorage)

			req := LoginRequest{
				Username: tt.username,
				Password: tt.password,
			}

			resp, err := service.Login(context.Background(), req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				// Service wraps storage errors in ErrorWithCode
				var govErr *governerrors.ErrorWithCode
				assert.ErrorAs(t, err, &govErr)
				assert.Equal(t, governerrors.CodeInternal, govErr.Code)
			}
		})
	}
}

func TestService_Login_Success(t *testing.T) {
	t.Run("successfully logs in with valid credentials", func(t *testing.T) {
		t.Parallel()

		mockStorage := storageMocks.NewMockStorage(t)
		mockUser, passwordHash := newMockUser(t, "testuser", "correctpass")
		mockStorage.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "testuser").Return(mockUser, passwordHash, nil)

		service := newTestService(t, mockStorage)

		req := LoginRequest{
			Username: "testuser",
			Password: "correctpass",
		}

		resp, err := service.Login(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Token, "JWT token should be generated")
		assert.NotNil(t, resp.User)
		assert.Equal(t, "testuser", resp.User.Username)
		// model.User doesn't have PasswordHash field - security by design
		assert.WithinDuration(t, time.Now().Add(testJWTExpiration), resp.ExpiresAt, 1*time.Second,
			"expiration should be approximately testJWTExpiration from now")

		// Verify token structure (following Uber: "verify invariants, not implementation")
		claims := &struct {
			ID       string
			Email    string
			Username string
			jwt.RegisteredClaims
		}{}
		token, err := jwt.ParseWithClaims(resp.Token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		require.NoError(t, err)
		assert.True(t, token.Valid, "JWT token should be valid")
		assert.Equal(t, "1", claims.ID)
		assert.Equal(t, "testuser@example.com", claims.Email)
		assert.Equal(t, "testuser", claims.Username)
	})
}

func TestService_Login_TokenExpiration(t *testing.T) {
	t.Run("generates token with correct expiration", func(t *testing.T) {
		t.Parallel()

		mockStorage := storageMocks.NewMockStorage(t)
		mockUser, passwordHash := newMockUser(t, "testuser", "password")
		mockStorage.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "testuser").Return(mockUser, passwordHash, nil)

		service := newTestService(t, mockStorage)

		req := LoginRequest{
			Username: "testuser",
			Password: "password",
		}

		resp, err := service.Login(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify expiration is set correctly
		expectedExpiry := time.Now().Add(testJWTExpiration)
		assert.WithinDuration(t, expectedExpiry, resp.ExpiresAt, 1*time.Second,
			"token expiration should match configured expiration")
	})
}

// Benchmark for Register operation
// Following Uber: "Benchmark before optimizing"
func BenchmarkService_Register(b *testing.B) {
	log := zap.NewNop().Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mockStorage := storageMocks.NewMockStorage(b)
		mockStorage.EXPECT().CheckUniqueness(mock.Anything, "testuser", "test@example.com").Return(false, false, nil)
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*model.User"), mock.AnythingOfType("string")).RunAndReturn(func(ctx context.Context, user *model.User, passwordHash string) (*model.User, error) {
			user.ID = 1
			return user, nil
		})
		service := NewAuthService(log, mockStorage, "test-secret", testJWTExpiration)
		b.StartTimer()

		req := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "SecurePass123!",
			FullName: "Test User",
		}
		_, _ = service.Register(context.Background(), req)
	}
}

// Benchmark for Login operation (hot path)
// Following Uber: "Benchmark realistic workflows"
func BenchmarkService_Login(b *testing.B) {
	log := zap.NewNop().Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hash, _ := password.HashPassword("correctpass")
		mockStorage := storageMocks.NewMockStorage(b)
		mockStorage.EXPECT().FindUserByUsernameWithPassword(mock.Anything, "testuser").Return(&model.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		}, hash, nil)
		service := NewAuthService(log, mockStorage, "test-secret", testJWTExpiration)
		b.StartTimer()

		req := LoginRequest{
			Username: "testuser",
			Password: "correctpass",
		}
		_, _ = service.Login(context.Background(), req)
	}
}
