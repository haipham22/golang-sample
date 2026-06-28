# Database Adapter Pattern - Sample Implementation

**Concrete example showing how to switch between GORM, pgx, and sqlc using adapter pattern.**

---

## Interface Definition (Database-Agnostic)

```go
// internal/storage/user/interface.go
package user

import (
    "context"
    "golang-sample/internal/model"
)

// Storage interface - database agnostic
type Storage interface {
    // CRUD operations
    CreateUser(ctx context.Context, user *model.User) error
    FindUserByID(ctx context.Context, id int64) (*model.User, error)
    FindUserByUsername(ctx context.Context, username string) (*model.User, error)
    FindUserByEmail(ctx context.Context, email string) (*model.User, error)
    UpdateUser(ctx context.Context, user *model.User) error
    DeleteUser(ctx context.Context, id int64) error
    
    // Business operations
    CheckUniqueness(ctx context.Context, username, email string) (usernameExists, emailExists bool, err error)
    CreateUserWithPassword(ctx context.Context, user *model.User, passwordHash string) (*model.User, error)
    FindUserByUsernameWithPassword(ctx context.Context, username string) (*model.User, string, error)
}
```

---

## GORM Implementation (Current)

```go
// internal/storage/user/gorm.go
package user

import (
    "context"
    "github.com/pkg/errors"
    "go.uber.org/zap"
    "gorm.io/gorm"
    
    "golang-sample/internal/model"
    "golang-sample/internal/orm"
)

type gormRepository struct {
    db  *gorm.DB
    log *zap.SugaredLogger
}

func NewGORM(db *gorm.DB, log *zap.SugaredLogger) Storage {
    return &gormRepository{db: db, log: log}
}

func (r *gormRepository) CreateUser(ctx context.Context, user *model.User) error {
    ormUser := modelToORM(user)
    
    if err := r.db.WithContext(ctx).Create(&ormUser).Error; err != nil {
        r.log.Errorf("Failed to create user: %v", err)
        return errors.Wrap(err, "failed to create user")
    }
    
    return nil
}

func (r *gormRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
    var ormUser orm.User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&ormUser).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, errors.Wrap(err, "failed to find user by username")
    }
    
    return ormToModel(ormUser), nil
}

func (r *gormRepository) CheckUniqueness(ctx context.Context, username, email string) (bool, bool, error) {
    type Result struct {
        UsernameCount int64
        EmailCount    int64
    }
    
    var result Result
    err := r.db.WithContext(ctx).Model(&orm.User{}).Select(`
        COUNT(CASE WHEN username = ? THEN 1 END) as username_count,
        COUNT(CASE WHEN email = ? THEN 1 END) as email_count
    `, username, email).Scan(&result).Error
    
    if err != nil {
        return false, false, errors.Wrap(err, "failed to check uniqueness")
    }
    
    return result.UsernameCount > 0, result.EmailCount > 0, nil
}

// ... other implementations
```

---

## PGX Implementation (Example)

```go
// internal/storage/user/pgx.go
package user

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pkg/errors"
    "go.uber.org/zap"
    
    "golang-sample/internal/model"
)

type pgxRepository struct {
    pool *pgxpool.Pool
    log  *zap.SugaredLogger
}

func NewPGX(pool *pgxpool.Pool, log *zap.SugaredLogger) Storage {
    return &pgxRepository{pool: pool, log: log}
}

func (r *pgxRepository) CreateUser(ctx context.Context, user *model.User) error {
    query := `
        INSERT INTO users (username, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
    
    err := r.pool.QueryRow(ctx, query,
        user.Username,
        user.Email,
        user.PasswordHash,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    
    if err != nil {
        r.log.Errorf("Failed to create user: %v", err)
        return errors.Wrap(err, "failed to create user")
    }
    
    return nil
}

func (r *pgxRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
    query := `
        SELECT id, username, email, created_at, updated_at
        FROM users
        WHERE username = $1
        AND deleted_at IS NULL
    `
    
    var user model.User
    err := r.pool.QueryRow(ctx, query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil
        }
        r.log.Errorf("Failed to find user: %v", err)
        return nil, errors.Wrap(err, "failed to find user by username")
    }
    
    return &user, nil
}

func (r *pgxRepository) CheckUniqueness(ctx context.Context, username, email string) (bool, bool, error) {
    query := `
        SELECT 
            COUNT(*) FILTER (WHERE username = $1) as username_count,
            COUNT(*) FILTER (WHERE email = $2) as email_count
        FROM users
        WHERE deleted_at IS NULL
    `
    
    var usernameCount, emailCount int64
    err := r.pool.QueryRow(ctx, query, username, email).Scan(&usernameCount, &emailCount)
    
    if err != nil {
        r.log.Errorf("Failed to check uniqueness: %v", err)
        return false, false, errors.Wrap(err, "failed to check uniqueness")
    }
    
    return usernameCount > 0, emailCount > 0, nil
}

func (r *pgxRepository) FindUserByUsernameWithPassword(ctx context.Context, username string) (*model.User, string, error) {
    query := `
        SELECT id, username, email, password_hash, created_at, updated_at
        FROM users
        WHERE username = $1
        AND deleted_at IS NULL
    `
    
    var user model.User
    var passwordHash string
    err := r.pool.QueryRow(ctx, query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &passwordHash,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, "", nil
        }
        return nil, "", errors.Wrap(err, "failed to find user with password")
    }
    
    return &user, passwordHash, nil
}

// ... other implementations
```

---

## Factory Pattern

```go
// internal/storage/user/factory.go
package user

import (
    "fmt"
    
    "gorm.io/gorm"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

type DatabaseDriver string

const (
    DriverGORM DatabaseDriver = "gorm"
    DriverPGX  DatabaseDriver = "pgx"
    DriverSQLC DatabaseDriver = "sqlc"
)

type Config struct {
    Driver DatabaseDriver
    GORM   *gorm.DB
    PGX    *pgxpool.Pool
    Log    *zap.SugaredLogger
}

func NewStorage(cfg Config) (Storage, error) {
    switch cfg.Driver {
    case DriverGORM:
        if cfg.GORM == nil {
            return nil, fmt.Errorf("gorm DB required for gorm driver")
        }
        return NewGORM(cfg.GORM, cfg.Log), nil
        
    case DriverPGX:
        if cfg.PGX == nil {
            return nil, fmt.Errorf("pgx pool required for pgx driver")
        }
        return NewPGX(cfg.PGX, cfg.Log), nil
        
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
    }
}
```

---

## Bootstrap Usage

```go
// internal/bootstrap/database.go
package bootstrap

import (
    "gorm.io/gorm"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
    
    "golang-sample/internal/storage/user"
)

// Option 1: Direct (current)
func NewUserStorage(db *gorm.DB, log *zap.SugaredLogger) user.Storage {
    return user.NewGORM(db, log)
}

// Option 2: Config-driven (flexible)
func NewUserStorageFromConfig(cfg user.Config) (user.Storage, error) {
    return user.NewStorage(cfg)
}

// Option 3: Environment-driven
func NewUserStorageFromEnv(driver string, db *gorm.DB, pool *pgxpool.Pool, log *zap.SugaredLogger) (user.Storage, error) {
    cfg := user.Config{
        Driver: user.DatabaseDriver(driver),
        GORM:   db,
        PGX:    pool,
        Log:    log,
    }
    return user.NewStorage(cfg)
}
```

---

## Service Layer (No Changes Needed)

```go
// internal/service/auth/service.go
package auth

type Service struct {
    userStorage user.Storage  // Interface - database agnostic
    log         *zap.SugaredLogger
}

func NewService(userStorage user.Storage, log *zap.SugaredLogger) *Service {
    return &Service{
        userStorage: userStorage,
        log:         log,
    }
}

// Service logic doesn't care about implementation
func (s *Service) LoginUser(ctx context.Context, username, password string) (string, error) {
    user, passwordHash, err := s.userStorage.FindUserByUsernameWithPassword(ctx, username)
    if err != nil {
        return "", err
    }
    
    if user == nil {
        return "", ErrInvalidCredentials
    }
    
    // ... rest of logic
}
```

---

## Configuration Example

```go
// config/config.go
package config

type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
    Driver string `mapstructure:"driver"` // "gorm" | "pgx" | "sqlc"
    DSN    string `mapstructure:"dsn"`
}

// Usage
func Load() (*Config, error) {
    cfg := &Config{
        Database: DatabaseConfig{
            Driver: "gorm",  // Can switch to "pgx"
            DSN:    "host=localhost user=postgres...",
        },
    }
    return cfg, nil
}
```

---

## Migration Path

### Phase 1: GORM (Current) ✅
```go
storage := user.NewGORM(db, log)
```

### Phase 2: Add PGX (Future)
```go
// Option A: Run both in parallel
gormStorage := user.NewGORM(db, log)
pgxStorage := user.NewPGX(pool, log)

// Use GORM for CRUD
// Use PGX for performance-critical queries

// Option B: Switch entirely
storage := user.NewStorage(user.Config{
    Driver: user.DriverPGX,
    PGX:    pool,
    Log:    log,
})
```

### Phase 3: Hybrid (Best of both)
```go
type HybridStorage struct {
    gorm *gormRepository
    pgx  *pgxRepository
}

func (h *HybridStorage) CreateUser(ctx context.Context, user *model.User) error {
    // Use GORM for complex operations
    return h.gorm.CreateUser(ctx, user)
}

func (h *HybridStorage) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
    // Use PGX for hot path
    return h.pgx.FindUserByUsername(ctx, username)
}
```

---

## Testing with Mocks

```go
// internal/storage/user/mock.go
package user

type MockStorage struct {
    Users map[string]*model.User
}

func NewMock() *MockStorage {
    return &MockStorage{
        Users: make(map[string]*model.User),
    }
}

func (m *MockStorage) CreateUser(ctx context.Context, user *model.User) error {
    m.Users[user.Username] = user
    return nil
}

func (m *MockStorage) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
    user, ok := m.Users[username]
    if !ok {
        return nil, nil
    }
    return user, nil
}

// Usage in tests
func TestService(t *testing.T) {
    mockStorage := user.NewMock()
    svc := auth.NewService(mockStorage, log)
    
    // Test with mock - no database needed
}
```

---

## Pros/Cons Summary

### ✅ Benefits
- **Easy switching** - Change config, recompile
- **Testing friendly** - Mock implementations easy
- **Clean architecture** - Service layer database-agnostic
- **Future-proof** - Add implementations without changing service layer

### ❌ Drawbacks
- **Abstraction overhead** - Extra function calls
- **Lowest common denominator** - Loses ORM-specific features
- **More code** - Multiple implementations to maintain
- **Upfront cost** - Time investment before benefit

---

## When to Use

✅ **Use when:**
- Planning to support multiple databases
- Need to A/B test implementations
- Building library/framework
- Uncertain about database choice

❌ **Skip when:**
- Only using one database (current case)
- Team small, time limited
- Leveraging ORM-specific features
- Performance critical (overhead matters)

---

## Recommendation for golang-sample

### ✅ **Keep current approach, add adapter later if needed**

**Current architecture already enables adapter pattern:**
- Storage interface exists ✅
- Service layer database-agnostic ✅
- Easy to add PGX implementation later ✅

**Timeline:**
- **Now:** Use GORM, optimize queries
- **Future:** Add PGX for hot paths if needed
- **Scale:** Consider full adapter pattern at scale

**Effort:** 6 hours (add PGX later) vs 16-24 hours (full adapter now)
