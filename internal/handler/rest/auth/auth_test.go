package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"golang-sample/pkg/models"
	"golang-sample/pkg/utils/password"
)

// mockStorage is a mock implementation of the storage interface
type mockStorage struct {
	user             *models.User
	findUserErr      error
	createUserErr    error
	createUser       *models.User
	isExistByResult  bool
	isExistByErr     error
	validateCalled   bool
	createCalled     bool
	passwordChecked  bool
	lastLoginUpdated bool
}

func (m *mockStorage) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	m.validateCalled = true
	if m.findUserErr != nil {
		return nil, m.findUserErr
	}
	return m.user, nil
}

func (m *mockStorage) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	m.createCalled = true
	m.createUser = user
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}
	if m.createUser != nil {
		return m.createUser, nil
	}
	// Return a copy with ID set
	user.ID = 1
	return user, nil
}

func (m *mockStorage) IsExistBy(field, value string) (bool, error) {
	m.validateCalled = true
	if m.isExistByErr != nil {
		return false, m.isExistByErr
	}
	return m.isExistByResult, nil
}

func (m *mockStorage) UpdateLastLogin(ctx context.Context, user *models.User, lastLogin time.Time) error {
	m.lastLoginUpdated = true
	return nil
}

func setupTestController(storage *mockStorage) *Controller {
	logger := zap.NewNop().Sugar()
	return &Controller{
		log:     logger,
		storage: storage,
	}
}

func TestPasswordVerification(t *testing.T) {
	t.Run("correct password verifies successfully", func(t *testing.T) {
		pwd := "TestPassword123!"
		hash, err := password.HashPassword(pwd)
		require.NoError(t, err)

		result := password.CheckPasswordHash(pwd, hash)
		assert.True(t, result, "Correct password should verify")
	})

	t.Run("incorrect password fails verification", func(t *testing.T) {
		pwd := "TestPassword123!"
		wrongPassword := "WrongPassword123!"
		hash, err := password.HashPassword(pwd)
		require.NoError(t, err)

		result := password.CheckPasswordHash(wrongPassword, hash)
		assert.False(t, result, "Incorrect password should not verify")
	})
}

func TestStorage_FindUserByUsername(t *testing.T) {
	t.Run("finds existing user", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hash",
		}
		storage := &mockStorage{user: expectedUser}

		user, err := storage.FindUserByUsername(context.Background(), "testuser")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		assert.True(t, storage.validateCalled)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		storage := &mockStorage{user: nil, findUserErr: expectedErr}

		user, err := storage.FindUserByUsername(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, expectedErr, err)
	})
}

func TestStorage_CreateUser(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		newUser := &models.User{
			Username:     "newuser",
			Email:        "new@example.com",
			PasswordHash: "hashedpassword",
		}
		storage := &mockStorage{createUser: newUser}

		createdUser, err := storage.CreateUser(context.Background(), newUser)

		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.True(t, storage.createCalled)
		assert.Equal(t, newUser.Username, storage.createUser.Username)
		assert.Equal(t, newUser.Email, storage.createUser.Email)
	})

	t.Run("returns error on creation failure", func(t *testing.T) {
		expectedErr := errors.New("database error")
		storage := &mockStorage{createUserErr: expectedErr}

		user := &models.User{
			Username: "testuser",
		}

		createdUser, err := storage.CreateUser(context.Background(), user)

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Equal(t, expectedErr, err)
	})
}

func TestStorage_IsExistBy(t *testing.T) {
	t.Run("field exists", func(t *testing.T) {
		storage := &mockStorage{isExistByResult: true}

		exists, err := storage.IsExistBy("username", "existinguser")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("field does not exist", func(t *testing.T) {
		storage := &mockStorage{isExistByResult: false}

		exists, err := storage.IsExistBy("username", "newuser")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns error on check failure", func(t *testing.T) {
		expectedErr := errors.New("database error")
		storage := &mockStorage{isExistByErr: expectedErr}

		exists, err := storage.IsExistBy("username", "testuser")

		assert.Error(t, err)
		assert.False(t, exists)
		assert.Equal(t, expectedErr, err)
	})
}

func TestStorage_UpdateLastLogin(t *testing.T) {
	t.Run("updates last login successfully", func(t *testing.T) {
		user := &models.User{
			ID:       1,
			Username: "testuser",
		}
		storage := &mockStorage{}

		err := storage.UpdateLastLogin(context.Background(), user, time.Now())

		assert.NoError(t, err)
		assert.True(t, storage.lastLoginUpdated)
	})
}

func TestPasswordHashing_Consistency(t *testing.T) {
	pwd := "ConsistentPassword123!"

	// Hash the same password twice
	hash1, err1 := password.HashPassword(pwd)
	hash2, err2 := password.HashPassword(pwd)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Hashes should be different (bcrypt uses salt)
	assert.NotEqual(t, hash1, hash2, "Hashes should be different due to salt")

	// But both should verify correctly
	assert.True(t, password.CheckPasswordHash(pwd, hash1))
	assert.True(t, password.CheckPasswordHash(pwd, hash2))
}

func TestController_Creation(t *testing.T) {
	storage := &mockStorage{}
	controller := setupTestController(storage)

	assert.NotNil(t, controller)
	assert.NotNil(t, controller.log)
	assert.NotNil(t, controller.storage)
	assert.Equal(t, storage, controller.storage)
}

// Benchmark tests
func BenchmarkPasswordHash(b *testing.B) {
	pwd := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = password.HashPassword(pwd)
	}
}

func BenchmarkPasswordCheck(b *testing.B) {
	pwd := "BenchmarkPassword123!"
	hash, err := password.HashPassword(pwd)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = password.CheckPasswordHash(pwd, hash)
	}
}
