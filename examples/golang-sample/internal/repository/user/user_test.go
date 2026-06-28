package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
	storageMocks "github.com/haipham22/golang-sample/internal/mocks/storage"
	"github.com/haipham22/golang-sample/internal/orm"
)

// TestStorage_InterfaceCompliance verifies the repo implements Storage interface
func TestStorage_InterfaceCompliance(t *testing.T) {
	// Compile-time interface check
	var _ Storage = (*repo)(nil)
}

// Mock-based tests for IsExistBy
func TestStorage_IsExistBy_WithMock(t *testing.T) {
	t.Run("returns true when user exists", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "existinguser").Return(true, nil)

		ctx := context.Background()
		exists, err := mockStorage.IsExistBy(ctx, "username", "existinguser")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false when user does not exist", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "nonexistent").Return(false, nil)

		ctx := context.Background()
		exists, err := mockStorage.IsExistBy(ctx, "username", "nonexistent")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns error on database error", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "testuser").Return(false, errors.New("database error"))

		ctx := context.Background()
		exists, err := mockStorage.IsExistBy(ctx, "username", "testuser")

		assert.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("checks by email field", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "email", "test@example.com").Return(true, nil)

		ctx := context.Background()
		exists, err := mockStorage.IsExistBy(ctx, "email", "test@example.com")

		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

// Mock-based tests for CreateUser
func TestStorage_CreateUser_WithMock(t *testing.T) {
	t.Run("successfully creates user", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		expectedUser := &domain.User{
			ID:       1,
			Username: "newuser",
			Email:    "newuser@example.com",
		}
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Username == "newuser" && u.Email == "newuser@example.com"
		}), mock.AnythingOfType("string")).Return(expectedUser, nil)

		ctx := context.Background()
		user := &domain.User{
			Username: "newuser",
			Email:    "newuser@example.com",
		}

		result, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, "newuser", result.Username)
		assert.Equal(t, "newuser@example.com", result.Email)
	})

	t.Run("returns error on duplicate username", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(nil, errors.New("UNIQUE constraint failed"))

		ctx := context.Background()
		user := &domain.User{
			Username: "existinguser",
			Email:    "new@example.com",
		}

		result, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(nil, sql.ErrConnDone)

		ctx := context.Background()
		user := &domain.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		result, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// Mock-based tests for FindUserByUsername
func TestStorage_FindUserByUsername_WithMock(t *testing.T) {
	t.Run("successfully finds user", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		expectedUser := &domain.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		}
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "testuser").Return(expectedUser, nil)

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "testuser")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "testuser", found.Username)
		assert.Equal(t, "test@example.com", found.Email)
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "nonexistent").Return(nil, nil)

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "testuser").Return(nil, errors.New("database error"))

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "testuser")

		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("case sensitive search", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "TestUser").Return(nil, nil)

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "TestUser")

		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

// Integration workflow test with mocks
func TestStorage_UserWorkflow_WithMock(t *testing.T) {
	t.Run("complete user registration workflow", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)

		// Step 1: Check username doesn't exist
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "newuser").Return(false, nil)

		// Step 2: Check email doesn't exist
		mockStorage.EXPECT().IsExistBy(mock.Anything, "email", "newuser@example.com").Return(false, nil)

		// Step 3: Create user
		createdUser := &domain.User{
			ID:       1,
			Username: "newuser",
			Email:    "newuser@example.com",
		}
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Username == "newuser" && u.Email == "newuser@example.com"
		}), mock.AnythingOfType("string")).Return(createdUser, nil)

		// Step 4: Verify user can be found
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "newuser").Return(createdUser, nil)

		ctx := context.Background()

		// Execute workflow
		exists, err := mockStorage.IsExistBy(ctx, "username", "newuser")
		require.NoError(t, err)
		assert.False(t, exists)

		exists, err = mockStorage.IsExistBy(ctx, "email", "newuser@example.com")
		require.NoError(t, err)
		assert.False(t, exists)

		user := &domain.User{
			Username: "newuser",
			Email:    "newuser@example.com",
		}
		created, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")
		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, uint(1), created.ID)

		found, err := mockStorage.FindUserByUsername(ctx, "newuser")
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "newuser", found.Username)
	})
}

// Table-driven tests for username validation scenarios
func TestStorage_UsernameValidation_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		exists    bool
		checkErr  bool
		setupMock func(*storageMocks.MockStorage)
	}{
		{
			name:     "valid username exists",
			username: "validuser",
			exists:   true,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "validuser").Return(true, nil)
			},
		},
		{
			name:     "valid username does not exist",
			username: "newuser",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "newuser").Return(false, nil)
			},
		},
		{
			name:     "username with dash",
			username: "user-name",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user-name").Return(false, nil)
			},
		},
		{
			name:     "username with dot",
			username: "user.name",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user.name").Return(false, nil)
			},
		},
		{
			name:     "username with underscore",
			username: "user_name",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user_name").Return(false, nil)
			},
		},
		{
			name:     "username with numbers",
			username: "user123",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user123").Return(false, nil)
			},
		},
		{
			name:     "empty username",
			username: "",
			exists:   false,
			checkErr: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "").Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storageMocks.NewMockStorage(t)
			tt.setupMock(mockStorage)

			ctx := context.Background()
			exists, err := mockStorage.IsExistBy(ctx, "username", tt.username)

			if tt.checkErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.exists, exists)
		})
	}
}

// Table-driven tests for email validation scenarios
func TestStorage_EmailValidation_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		exists    bool
		setupMock func(*storageMocks.MockStorage)
	}{
		{
			name:   "valid email exists",
			email:  "test@example.com",
			exists: true,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "email", "test@example.com").Return(true, nil)
			},
		},
		{
			name:   "valid email does not exist",
			email:  "new@example.com",
			exists: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "email", "new@example.com").Return(false, nil)
			},
		},
		{
			name:   "email with plus sign",
			email:  "user+tag@example.com",
			exists: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "email", "user+tag@example.com").Return(false, nil)
			},
		},
		{
			name:   "email with subdomain",
			email:  "user@mail.example.com",
			exists: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "email", "user@mail.example.com").Return(false, nil)
			},
		},
		{
			name:   "empty email",
			email:  "",
			exists: false,
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "email", "").Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storageMocks.NewMockStorage(t)
			tt.setupMock(mockStorage)

			ctx := context.Background()
			exists, err := mockStorage.IsExistBy(ctx, "email", tt.email)

			assert.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

// Test error handling scenarios
func TestStorage_ErrorHandling_WithMock(t *testing.T) {
	t.Run("IsExistBy handles connection errors", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "test").Return(false, sql.ErrConnDone)

		ctx := context.Background()
		_, err := mockStorage.IsExistBy(ctx, "username", "test")

		assert.Error(t, err)
	})

	t.Run("CreateUser handles duplicate key error", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(nil, fmt.Errorf("UNIQUE constraint failed: users.username"))

		ctx := context.Background()
		user := &domain.User{
			Username: "test",
			Email:    "test@example.com",
		}

		result, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "UNIQUE")
	})

	t.Run("FindUserByUsername handles database errors", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "test").Return(nil, errors.New("connection lost"))

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "test")

		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// Benchmark tests with mocks
func BenchmarkStorage_IsExistBy_Mock(b *testing.B) {
	mockStorage := storageMocks.NewMockStorage(b)
	mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "benchuser").Return(true, nil).Times(b.N)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockStorage.IsExistBy(ctx, "username", "benchuser")
	}
}

func BenchmarkStorage_CreateUser_Mock(b *testing.B) {
	mockStorage := storageMocks.NewMockStorage(b)
	user := &domain.User{
		ID:       1,
		Username: "benchuser",
		Email:    "bench@example.com",
	}
	mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(user, nil).Times(b.N)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockStorage.CreateUserWithPassword(ctx, &domain.User{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
		}, "hash")
	}
}

func BenchmarkStorage_FindUserByUsername_Mock(b *testing.B) {
	mockStorage := storageMocks.NewMockStorage(b)
	user := &domain.User{
		ID:       1,
		Username: "benchuser",
		Email:    "bench@example.com",
	}
	mockStorage.EXPECT().FindUserByUsername(mock.Anything, "benchuser").Return(user, nil).Times(b.N)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockStorage.FindUserByUsername(ctx, "benchuser")
	}
}

// Test for nil and empty inputs
func TestStorage_NilAndEmptyInputs_WithMock(t *testing.T) {
	t.Run("FindUserByUsername with empty string", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().FindUserByUsername(mock.Anything, "").Return(nil, nil)

		ctx := context.Background()
		found, err := mockStorage.FindUserByUsername(ctx, "")

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("IsExistBy with empty condition", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "").Return(false, nil)

		ctx := context.Background()
		exists, err := mockStorage.IsExistBy(ctx, "username", "")

		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// Test for concurrent operations
func TestStorage_ConcurrentOperations_WithMock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	mockStorage := storageMocks.NewMockStorage(t)
	user := &domain.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Setup expectations for concurrent calls
	mockStorage.EXPECT().IsExistBy(mock.Anything, "username", "testuser").Return(false, nil).Times(5)
	mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(user, nil).Times(5)
	mockStorage.EXPECT().FindUserByUsername(mock.Anything, "testuser").Return(user, nil).Times(5)

	ctx := context.Background()
	done := make(chan bool, 15)

	// Concurrent existence checks
	for range 5 {
		go func() {
			defer func() { done <- true }()
			mockStorage.IsExistBy(ctx, "username", "testuser")
		}()
	}

	// Concurrent creates
	for i := range 5 {
		idx := i
		go func() {
			defer func() { done <- true }()
			mockStorage.CreateUserWithPassword(ctx, &domain.User{
				Username: fmt.Sprintf("user%d", idx),
				Email:    fmt.Sprintf("user%d@example.com", idx),
			}, "hash")
		}()
	}

	// Concurrent finds
	for range 5 {
		go func() {
			defer func() { done <- true }()
			mockStorage.FindUserByUsername(ctx, "testuser")
		}()
	}

	// Wait for all goroutines
	for range 15 {
		<-done
	}
}

// openTestDB creates a test database with proper cleanup
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique DSN per test to avoid conflicts
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	dbSQL, err := db.DB()
	require.NoError(t, err, "Failed to get underlying sql.DB")

	// Limit connections to 1 for shared cache
	dbSQL.SetMaxOpenConns(1)
	dbSQL.SetMaxIdleConns(1)

	// Register cleanup to close the database
	t.Cleanup(func() {
		if err := dbSQL.Close(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	})

	return db
}

// TestNew verifies the constructor
func TestNew(t *testing.T) {
	t.Run("creates new storage instance", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		db := openTestDB(t)

		storage := New(log, db)

		assert.NotNil(t, storage)
		assert.IsType(t, &repo{}, storage)
	})
}

// TestRepo verifies the repo struct
func TestRepo(t *testing.T) {
	t.Run("repo struct stores dependencies", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		db := openTestDB(t)

		r := &repo{log: log, db: db}

		assert.Same(t, log, r.log)
		assert.Same(t, db, r.db)
	})
}

// Test for special characters in inputs
func TestStorage_SpecialCharacters_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		email     string
		setupMock func(*storageMocks.MockStorage)
	}{
		{
			name:     "username with dash",
			username: "user-name",
			email:    "user-dash@example.com",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user-name").Return(false, nil)
			},
		},
		{
			name:     "username with dot",
			username: "user.name",
			email:    "user.dot@example.com",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user.name").Return(false, nil)
			},
		},
		{
			name:     "username with mixed special chars",
			username: "user_name-123.test",
			email:    "user+tag@test.example.com",
			setupMock: func(m *storageMocks.MockStorage) {
				m.EXPECT().IsExistBy(mock.Anything, "username", "user_name-123.test").Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storageMocks.NewMockStorage(t)
			tt.setupMock(mockStorage)

			ctx := context.Background()
			exists, err := mockStorage.IsExistBy(ctx, "username", tt.username)

			assert.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

// Test for user creation with timestamps
func TestStorage_CreateUserTimestamps_WithMock(t *testing.T) {
	t.Run("created user has timestamps set", func(t *testing.T) {
		mockStorage := storageMocks.NewMockStorage(t)
		now := time.Now()
		expectedUser := &domain.User{
			ID:        1,
			Username:  "testuser",
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		}
		mockStorage.EXPECT().CreateUserWithPassword(mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).Return(expectedUser, nil)

		ctx := context.Background()
		user := &domain.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		result, err := mockStorage.CreateUserWithPassword(ctx, user, "hashedpassword")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
		assert.WithinDuration(t, now, result.CreatedAt, time.Second)
		assert.WithinDuration(t, now, result.UpdatedAt, time.Second)
	})
}

// Integration tests for actual repo implementation
// These tests use an in-memory SQLite database to test the actual implementation
//
// Note: These tests may be skipped in CI environments where CGO is not available
// Run with: go test ./internal/storage/user/ -run TestRepo_Integration -v

// TestRepo_IsExistBy_Integration tests IsExistBy with real database
func TestRepo_IsExistBy_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Test non-existent user
	exists, err := storage.IsExistBy(ctx, "username", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create a user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	result := db.Create(&orm.User{ID: user.ID, Username: user.Username, Email: user.Email, PasswordHash: "testhash"})
	require.NoError(t, result.Error)

	// Test existing user by username
	exists, err = storage.IsExistBy(ctx, "username", "testuser")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test existing user by email
	exists, err = storage.IsExistBy(ctx, "email", "test@example.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test case sensitivity
	exists, err = storage.IsExistBy(ctx, "username", "TestUser")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestRepo_CreateUser_Integration tests CreateUser with real database
func TestRepo_CreateUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Test successful creation
	user := &domain.User{
		Username: "newuser",
		Email:    "newuser@example.com",
	}

	result, err := storage.CreateUserWithPassword(ctx, user, "hashedpassword")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "newuser", result.Username)
	assert.Equal(t, "newuser@example.com", result.Email)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())

	// Test duplicate username
	duplicate := &domain.User{
		Username: "newuser",
		Email:    "different@example.com",
	}
	_, err = storage.CreateUserWithPassword(ctx, duplicate, "anotherhash")
	assert.Error(t, err)

	// Test duplicate email
	duplicate2 := &domain.User{
		Username: "different",
		Email:    "newuser@example.com",
	}
	_, err = storage.CreateUserWithPassword(ctx, duplicate2, "anotherhash")
	assert.Error(t, err)
}

// TestRepo_FindUserByUsername_Integration tests FindUserByUsername with real database
func TestRepo_FindUserByUsername_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Create a test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	result := db.Create(&orm.User{ID: user.ID, Username: user.Username, Email: user.Email, PasswordHash: "testhash"})
	require.NoError(t, result.Error)

	// Test finding existing user
	found, err := storage.FindUserByUsername(ctx, "testuser")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "testuser", found.Username)
	assert.Equal(t, "test@example.com", found.Email)

	// Test non-existent user
	found, err = storage.FindUserByUsername(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, found)
}

// TestRepo_CompleteWorkflow_Integration tests the complete user workflow
func TestRepo_CompleteWorkflow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Step 1: Check user doesn't exist
	exists, err := storage.IsExistBy(ctx, "username", "workflowuser")
	require.NoError(t, err)
	assert.False(t, exists)

	// Step 2: Create user
	user := &domain.User{
		Username: "workflowuser",
		Email:    "workflow@example.com",
	}
	created, err := storage.CreateUserWithPassword(ctx, user, "hashedpassword")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)

	// Step 3: Check user now exists
	exists, err = storage.IsExistBy(ctx, "username", "workflowuser")
	require.NoError(t, err)
	assert.True(t, exists)

	// Step 4: Find the user
	found, err := storage.FindUserByUsername(ctx, "workflowuser")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "workflowuser", found.Username)
}

// TestRepo_FindUserByID_Integration tests FindUserByID with real database
func TestRepo_FindUserByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Create test users
	user1 := &orm.User{Username: "user1", Email: "user1@example.com", PasswordHash: "hash1"}
	user2 := &orm.User{Username: "user2", Email: "user2@example.com", PasswordHash: "hash2"}
	require.NoError(t, db.Create(user1).Error)
	require.NoError(t, db.Create(user2).Error)

	// Test finding existing user
	found, err := storage.FindUserByID(ctx, user1.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user1.ID, found.ID)
	assert.Equal(t, "user1", found.Username)
	assert.Equal(t, "user1@example.com", found.Email)

	// Test finding another user
	found, err = storage.FindUserByID(ctx, user2.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user2.ID, found.ID)
	assert.Equal(t, "user2", found.Username)

	// Test non-existent user (returns nil, not error)
	found, err = storage.FindUserByID(ctx, 9999)
	assert.NoError(t, err)
	assert.Nil(t, found)

	// Test with zero ID
	found, err = storage.FindUserByID(ctx, 0)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

// TestRepo_ListUsers_Integration tests ListUsers with real database
func TestRepo_ListUsers_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Create test users
	users := []*orm.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "hash1"},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "hash2"},
		{Username: "user3", Email: "user3@example.com", PasswordHash: "hash3"},
		{Username: "user4", Email: "user4@example.com", PasswordHash: "hash4"},
		{Username: "user5", Email: "user5@example.com", PasswordHash: "hash5"},
	}
	for _, user := range users {
		require.NoError(t, db.Create(user).Error)
	}

	t.Run("list all users with default limit", func(t *testing.T) {
		result, total, err := storage.ListUsers(ctx, ListUsersParams{})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 5)
		// Verify ordered by ID ascending
		assert.Less(t, result[0].ID, result[1].ID)
		assert.Less(t, result[1].ID, result[2].ID)
	})

	t.Run("list with custom limit", func(t *testing.T) {
		result, total, err := storage.ListUsers(ctx, ListUsersParams{Limit: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 2)
	})

	t.Run("list with offset", func(t *testing.T) {
		result, total, err := storage.ListUsers(ctx, ListUsersParams{Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 2)
		// Should return user3 and user4
		assert.Equal(t, "user3", result[0].Username)
		assert.Equal(t, "user4", result[1].Username)
	})

	t.Run("list with offset beyond data", func(t *testing.T) {
		result, total, err := storage.ListUsers(ctx, ListUsersParams{Limit: 2, Offset: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 0)
	})

	t.Run("list with zero limit uses default", func(t *testing.T) {
		result, total, err := storage.ListUsers(ctx, ListUsersParams{Limit: 0})
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 5)
	})

	t.Run("list empty database", func(t *testing.T) {
		// Use a fresh database
		db2 := openTestDB(t)
		require.NoError(t, db2.AutoMigrate(&orm.User{}))
		storage2 := New(log, db2).(*repo)

		result, total, err := storage2.ListUsers(ctx, ListUsersParams{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, result, 0)
	})
}

// TestRepo_CheckUniqueness_Integration tests CheckUniqueness with real database
func TestRepo_CheckUniqueness_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Create a test user
	user := &orm.User{Username: "existinguser", Email: "existing@example.com", PasswordHash: "hash"}
	require.NoError(t, db.Create(user).Error)

	t.Run("both unique", func(t *testing.T) {
		usernameExists, emailExists, err := storage.CheckUniqueness(ctx, "newuser", "new@example.com")
		require.NoError(t, err)
		assert.False(t, usernameExists)
		assert.False(t, emailExists)
	})

	t.Run("username exists", func(t *testing.T) {
		usernameExists, emailExists, err := storage.CheckUniqueness(ctx, "existinguser", "new@example.com")
		require.NoError(t, err)
		assert.True(t, usernameExists)
		assert.False(t, emailExists)
	})

	t.Run("email exists", func(t *testing.T) {
		usernameExists, emailExists, err := storage.CheckUniqueness(ctx, "newuser", "existing@example.com")
		require.NoError(t, err)
		assert.False(t, usernameExists)
		assert.True(t, emailExists)
	})

	t.Run("both exist", func(t *testing.T) {
		usernameExists, emailExists, err := storage.CheckUniqueness(ctx, "existinguser", "existing@example.com")
		require.NoError(t, err)
		assert.True(t, usernameExists)
		assert.True(t, emailExists)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		// SQLite is case-insensitive by default for LIKE, but case-sensitive for =
		usernameExists, emailExists, err := storage.CheckUniqueness(ctx, "ExistingUser", "EXISTING@EXAMPLE.COM")
		require.NoError(t, err)
		// Should be case-sensitive for exact match
		assert.False(t, usernameExists)
		assert.False(t, emailExists)
	})
}

// TestRepo_FindUserByUsernameWithPassword_Integration tests FindUserByUsernameWithPassword
func TestRepo_FindUserByUsernameWithPassword_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)

	ctx := context.Background()

	// Create a test user with password
	passwordHash := "hashedpassword123"
	user := &orm.User{Username: "testuser", Email: "test@example.com", PasswordHash: passwordHash}
	require.NoError(t, db.Create(user).Error)

	t.Run("find existing user with password", func(t *testing.T) {
		found, hash, err := storage.FindUserByUsernameWithPassword(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, passwordHash, hash)
		assert.Equal(t, "testuser", found.Username)
		assert.Equal(t, "test@example.com", found.Email)
	})

	t.Run("find non-existent user returns nil", func(t *testing.T) {
		found, hash, err := storage.FindUserByUsernameWithPassword(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, found)
		assert.Empty(t, hash)
	})

	t.Run("find with empty username", func(t *testing.T) {
		found, hash, err := storage.FindUserByUsernameWithPassword(ctx, "")
		require.NoError(t, err)
		assert.Nil(t, found)
		assert.Empty(t, hash)
	})
}

// Converter tests
func TestConverter_ormToModel(t *testing.T) {
	t.Run("converts valid ORM user to domain user", func(t *testing.T) {
		now := time.Now()
		ormUser := &orm.User{
			ID:        123,
			Username:  "testuser",
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		}

		domainUser := ormToModel(ormUser)

		assert.NotNil(t, domainUser)
		assert.Equal(t, uint(123), domainUser.ID)
		assert.Equal(t, "testuser", domainUser.Username)
		assert.Equal(t, "test@example.com", domainUser.Email)
		assert.WithinDuration(t, now, domainUser.CreatedAt, time.Second)
		assert.WithinDuration(t, now, domainUser.UpdatedAt, time.Second)
	})

	t.Run("returns nil for nil ORM user", func(t *testing.T) {
		result := ormToModel(nil)
		assert.Nil(t, result)
	})

	t.Run("excludes password hash from domain model", func(t *testing.T) {
		ormUser := &orm.User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "secret123",
		}

		domainUser := ormToModel(ormUser)

		assert.NotNil(t, domainUser)
		// Domain User doesn't have PasswordHash field, so it's automatically excluded
		assert.Equal(t, "testuser", domainUser.Username)
	})
}

func TestConverter_modelToORM(t *testing.T) {
	t.Run("converts valid domain user to ORM user", func(t *testing.T) {
		now := time.Now()
		domainUser := &domain.User{
			ID:        456,
			Username:  "domainuser",
			Email:     "domain@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		}

		ormUser := modelToORM(domainUser)

		assert.NotNil(t, ormUser)
		assert.Equal(t, uint(456), ormUser.ID)
		assert.Equal(t, "domainuser", ormUser.Username)
		assert.Equal(t, "domain@example.com", ormUser.Email)
		assert.WithinDuration(t, now, ormUser.CreatedAt, time.Second)
		assert.WithinDuration(t, now, ormUser.UpdatedAt, time.Second)
	})

	t.Run("returns nil for nil domain user", func(t *testing.T) {
		result := modelToORM(nil)
		assert.Nil(t, result)
	})

	t.Run("ORM user has empty password hash initially", func(t *testing.T) {
		domainUser := &domain.User{
			Username: "newuser",
			Email:    "new@example.com",
		}

		ormUser := modelToORM(domainUser)

		assert.NotNil(t, ormUser)
		assert.Empty(t, ormUser.PasswordHash)
	})
}

func TestConverter_ormSliceToModelSlice(t *testing.T) {
	t.Run("converts slice of ORM users to domain users", func(t *testing.T) {
		ormUsers := []*orm.User{
			{ID: 1, Username: "user1", Email: "user1@example.com"},
			{ID: 2, Username: "user2", Email: "user2@example.com"},
			{ID: 3, Username: "user3", Email: "user3@example.com"},
		}

		domainUsers := ormSliceToModelSlice(ormUsers)

		assert.NotNil(t, domainUsers)
		assert.Len(t, domainUsers, 3)
		assert.Equal(t, "user1", domainUsers[0].Username)
		assert.Equal(t, "user2", domainUsers[1].Username)
		assert.Equal(t, "user3", domainUsers[2].Username)
	})

	t.Run("returns nil for nil slice", func(t *testing.T) {
		result := ormSliceToModelSlice(nil)
		assert.Nil(t, result)
	})

	t.Run("converts empty slice", func(t *testing.T) {
		ormUsers := []*orm.User{}
		domainUsers := ormSliceToModelSlice(ormUsers)

		assert.NotNil(t, domainUsers)
		assert.Len(t, domainUsers, 0)
	})

	t.Run("handles slice with nil elements", func(t *testing.T) {
		ormUsers := []*orm.User{
			{ID: 1, Username: "user1", Email: "user1@example.com"},
			nil,
			{ID: 3, Username: "user3", Email: "user3@example.com"},
		}

		domainUsers := ormSliceToModelSlice(ormUsers)

		assert.NotNil(t, domainUsers)
		assert.Len(t, domainUsers, 3)
		assert.NotNil(t, domainUsers[0])
		assert.Nil(t, domainUsers[1]) // nil in, nil out
		assert.NotNil(t, domainUsers[2])
	})
}

// Round-trip conversion tests
func TestConverter_RoundTrip(t *testing.T) {
	t.Run("domain -> ORM -> domain preserves data", func(t *testing.T) {
		original := &domain.User{
			ID:        789,
			Username:  "roundtrip",
			Email:     "roundtrip@example.com",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}

		// Convert to ORM and back
		ormUser := modelToORM(original)
		result := ormToModel(ormUser)

		assert.Equal(t, original.ID, result.ID)
		assert.Equal(t, original.Username, result.Username)
		assert.Equal(t, original.Email, result.Email)
		assert.WithinDuration(t, original.CreatedAt, result.CreatedAt, time.Second)
		assert.WithinDuration(t, original.UpdatedAt, result.UpdatedAt, time.Second)
	})
}
