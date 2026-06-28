package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/orm"
)

// BenchmarkFindUserByID benchmarks finding a user by ID
func BenchmarkFindUserByID(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:bench_find_by_id?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		b.Fatal(err)
	}

	// Create test users
	for i := 0; i < 100; i++ {
		user := &orm.User{
			Username:     fmt.Sprintf("user%d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hash",
		}
		if err := db.Create(user).Error; err != nil {
			b.Fatal(err)
		}
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between different IDs to test cache effects
		id := uint(i%100 + 1)
		_, _ = storage.FindUserByID(ctx, id)
	}
}

// BenchmarkListUsers benchmarks listing users with pagination
func BenchmarkListUsers(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:bench_list?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		b.Fatal(err)
	}

	// Create test users
	for i := 0; i < 100; i++ {
		user := &orm.User{
			Username:     fmt.Sprintf("user%d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hash",
		}
		if err := db.Create(user).Error; err != nil {
			b.Fatal(err)
		}
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = storage.ListUsers(ctx, ListUsersParams{Limit: 10, Offset: i % 10})
	}
}

// BenchmarkCheckUniqueness benchmarks uniqueness checking
func BenchmarkCheckUniqueness(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:bench_uniqueness?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		b.Fatal(err)
	}

	// Create a test user
	user := &orm.User{
		Username:     "existinguser",
		Email:        "existing@example.com",
		PasswordHash: "hash",
	}
	if err := db.Create(user).Error; err != nil {
		b.Fatal(err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("user%d@example.com", i)
		_, _, _ = storage.CheckUniqueness(ctx, username, email)
	}
}

// BenchmarkCreateUser benchmarks user creation
func BenchmarkCreateUser(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:bench_create?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		b.Fatal(err)
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &domain.User{
			Username: fmt.Sprintf("benchuser%d", i),
			Email:    fmt.Sprintf("benchuser%d@example.com", i),
		}
		_, _ = storage.CreateUserWithPassword(ctx, user, "hashedpassword")
	}
}

// BenchmarkFindUserByUsername benchmarks finding user by username
func BenchmarkFindUserByUsername(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:bench_find_username?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()

	if err := db.AutoMigrate(&orm.User{}); err != nil {
		b.Fatal(err)
	}

	// Create test users
	for i := 0; i < 100; i++ {
		user := &orm.User{
			Username:     fmt.Sprintf("user%d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hash",
		}
		if err := db.Create(user).Error; err != nil {
			b.Fatal(err)
		}
	}

	log := zap.NewNop().Sugar()
	storage := New(log, db).(*repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("user%d", i%100)
		_, _ = storage.FindUserByUsername(ctx, username)
	}
}

// BenchmarkConverter_ormToModel benchmarks ORM to domain conversion
func BenchmarkConverter_ormToModel(b *testing.B) {
	ormUser := &orm.User{
		ID:        123,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ormToModel(ormUser)
	}
}

// BenchmarkConverter_modelToORM benchmarks domain to ORM conversion
func BenchmarkConverter_modelToORM(b *testing.B) {
	domainUser := &domain.User{
		ID:        456,
		Username:  "domainuser",
		Email:     "domain@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = modelToORM(domainUser)
	}
}

// BenchmarkConverter_ormSliceToModelSlice benchmarks slice conversion
func BenchmarkConverter_ormSliceToModelSlice(b *testing.B) {
	ormUsers := make([]*orm.User, 100)
	for i := 0; i < 100; i++ {
		ormUsers[i] = &orm.User{
			ID:       uint(i + 1),
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ormSliceToModelSlice(ormUsers)
	}
}
