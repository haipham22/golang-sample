# Go Database & GORM Rules

**Best practices for database operations, GORM usage, query optimization, and transaction handling.**

---

## GORM Context Usage

**ALWAYS pass context to database operations:**
```go
// GOOD
func (r *Repository) FindUser(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    return &user, err
}

// BAD
func (r *Repository) FindUser(id int64) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    return &user, err
}
```

**Context rules:**
- ✅ Use `WithContext(ctx)` for all queries
- ✅ Context enables cancellation and timeouts
- ✅ Pass context through call chain
- ❌ NEVER use database without context

---

## Transaction Handling

**Use transactions for multi-step operations:**
```go
// GOOD - Transaction with automatic rollback on error
func (r *Repository) CreateUserWithProfile(ctx context.Context, user User, profile Profile) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&user).Error; err != nil {
            return err  // Auto-rollback
        }
        
        profile.UserID = user.ID
        if err := tx.Create(&profile).Error; err != nil {
            return err  // Auto-rollback
        }
        
        return nil  // Auto-commit
    })
}

// BAD - No transaction (data inconsistency risk)
func (r *Repository) CreateUserWithProfile(ctx context.Context, user User, profile Profile) error {
    if err := r.db.Create(&user).Error; err != nil {
        return err
    }
    
    profile.UserID = user.ID
    if err := r.db.Create(&profile).Error; err != nil {
        return err  // User created but profile failed - orphaned record
    }
    
    return nil
}
```

---

## Transaction-Aware Repositories

**Create repositories that can work with transactions:**
```go
// GOOD - Transaction-aware interface
type Repository interface {
    // Standard operations
    CreateUser(ctx context.Context, user *User) error
    
    // Transaction-aware operations
    CreateUserInTx(ctx context.Context, tx *gorm.DB, user *User) error
    UpdateUserInTx(ctx context.Context, tx *gorm.DB, user *User) error
}

// GOOD - Implementation
type repository struct {
    db *gorm.DB
}

func New(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) CreateUserInTx(ctx context.Context, tx *gorm.DB, user *User) error {
    return tx.WithContext(ctx).Create(user).Error
}

func (r *repository) UpdateUserInTx(ctx context.Context, tx *gorm.DB, user *User) error {
    return tx.WithContext(ctx).Save(user).Error
}
```

**Service layer orchestrates transactions:**
```go
// GOOD - Service controls transaction
func (s *Service) CreateUserWithProfile(ctx context.Context, req CreateUserRequest) error {
    return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
        // Create user
        user := &User{Username: req.Username, Email: req.Email}
        if err := s.repo.CreateUserInTx(ctx, tx, user); err != nil {
            return err
        }
        
        // Create profile with user ID
        profile := &Profile{
            UserID: user.ID,
            Bio:     req.Bio,
        }
        if err := s.profileRepo.CreateProfileInTx(ctx, tx, profile); err != nil {
            return err
        }
        
        // Create settings
        settings := &Settings{
            UserID:  user.ID,
            Theme:   req.Theme,
        }
        return s.settingsRepo.CreateSettingsInTx(ctx, tx, settings)
    })
}
```

---

## Dynamic Repository Creation

**Create repositories dynamically within transactions:**
```go
// GOOD - Factory pattern for transaction repos
func (r *repository) WithTx(tx *gorm.DB) *repository {
    return &repository{db: tx}
}

// Usage in service
func (s *Service) ComplexOperation(ctx context.Context) error {
    return s.db.Transaction(ctx, func(tx *gorm.DB) error {
        // Create transaction-aware repositories
        userRepo := s.userRepo.WithTx(tx)
        profileRepo := s.profileRepo.WithTx(tx)
        settingsRepo := s.settingsRepo.WithTx(tx)
        
        // Use repos within transaction
        if err := userRepo.CreateUserInTx(ctx, user); err != nil {
            return err
        }
        
        if err := profileRepo.CreateProfileInTx(ctx, profile); err != nil {
            return err
        }
        
        return settingsRepo.CreateSettingsInTx(ctx, settings)
    })
}
```

**Generic transaction helper:**
```go
// GOOD - Generic transaction runner
func (r *repository) Transaction(ctx context.Context, fn func(repo Repository) error) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txRepo := r.WithTx(tx)
        return fn(txRepo)
    })
}

// Usage
err := repo.Transaction(ctx, func(repo Repository) error {
    if err := repo.CreateUser(ctx, user1); err != nil {
        return err
    }
    if err := repo.CreateUser(ctx, user2); err != nil {
        return err
    }
    return nil
})
```

---

## Transaction Best Practices

### When to Use Transactions

✅ **Use transactions for:**
- Multi-step database operations
- Related data consistency required
- Atomic updates across multiple tables
- Complex business logic

❌ **Don't use transactions for:**
- Single database operation
- Read-only operations
- Independent operations
- Simple CRUD

### Transaction Rules

**1. Keep transactions short:**
```go
// GOOD - Minimal work in transaction
func (r *Repository) TransferBalance(ctx context.Context, fromID, toID int64, amount float64) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Only database operations
        if err := tx.Model(&User{}).Where("id = ?", fromID).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }
        if err := tx.Model(&User{}).Where("id = ?", toID).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }
        return nil
    })
}

// BAD - Long-running transaction
func (r *Repository) ProcessUser(ctx context.Context, user User) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        tx.Create(&user)
        
        // External API call (should not be in transaction)
        sendEmail(user.Email)  
        
        // File processing
        processAvatar(user.Avatar)
        
        return nil
    })
}
```

**2. Manual transaction control (when needed):**
```go
// GOOD - Manual transaction with explicit control
func (r *Repository) ComplexOperation(ctx context.Context) error {
    tx := r.db.WithContext(ctx).Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.Commit().Error; err != nil {
        return err
    }
    
    return nil
}
```

**3. Nested transactions (savepoints):**
```go
// GOOD - Use savepoints for nested logic
func (r *Repository) CreateUserWithOrders(ctx context.Context, user User, orders []Order) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        
        // Savepoint for orders
        if err := tx.SavePoint("orders").Error; err != nil {
            return err
        }
        
        for _, order := range orders {
            if err := tx.Create(&order).Error; err != nil {
                tx.RollbackTo("orders")
                return err
            }
        }
        
        return tx.ReleaseSavePoint("orders").Error
    })
}
```

**4. Transaction with isolation level:**
```go
// GOOD - Set isolation level when needed
func (r *Repository) SensitiveOperation(ctx context.Context) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        tx = tx.Begin(&gorm.Config{
            SkipDefaultTransaction: true,
        })
        
        // Use SERIALIZABLE for critical operations
        tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
        
        // ... operations
        
        return tx.Commit().Error
    })
}
```

**5. Read-only transactions:**
```go
// GOOD - Read-only transaction for consistent reads
func (r *Repository) GetUserAndProfile(ctx context.Context, userID int64) (*User, *Profile, error) {
    var user User
    var profile Profile
    
    err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
            return err
        }
        if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
            return err
        }
        return nil
    })
    
    if err != nil {
        return nil, nil, err
    }
    
    return &user, &profile, nil
}
```

### Transaction Patterns

**Repository pattern with transactions:**
```go
// GOOD - Repository accepts transaction
type Repository struct {
    db *gorm.DB
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) CreateUserInTx(ctx context.Context, tx *gorm.DB, user *User) error {
    return tx.Create(user).Error
}

// Service layer orchestrates transaction
func (s *Service) CreateUserWithProfile(ctx context.Context, req CreateUserRequest) error {
    return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
        user := &User{Username: req.Username}
        if err := s.repo.CreateUserInTx(ctx, tx, user); err != nil {
            return err
        }
        
        profile := &Profile{UserID: user.ID, Bio: req.Bio}
        if err := s.repo.CreateProfileInTx(ctx, tx, profile); err != nil {
            return err
        }
        
        return nil
    })
}
```

**Transaction helper:**
```go
// GOOD - Transaction helper
func (r *Repository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return fn(tx)
    })
}

// Usage
err := repo.Transaction(ctx, func(tx *gorm.DB) error {
    // Multiple operations
    return nil
})
```

### Transaction Error Handling

**Always check transaction errors:**
```go
// GOOD - Check commit error
err := r.db.Transaction(func(tx *gorm.DB) error {
    return tx.Create(&user).Error
})
if err != nil {
    return fmt.Errorf("transaction failed: %w", err)
}

// GOOD - Manual transaction with rollback
tx := r.db.Begin()
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}
if err := tx.Commit().Error; err != nil {
    return fmt.Errorf("commit failed: %w", err)
}
```

### Transaction Isolation Levels

**Choose appropriate isolation level:**

| Level | Use Case | Performance |
|-------|----------|-------------|
| **Read Uncommitted** | Rare, special cases | Fastest |
| **Read Committed** | Default for most operations | Fast |
| **Repeatable Read** | Consistent reads across transaction | Medium |
| **Serializable** | Critical operations, no conflicts allowed | Slowest |

```go
// GOOD - Set isolation level
tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
```

---

## Query Optimization

**Preload associations to avoid N+1:**
```go
// GOOD - Eager loading
users, err := r.db.Preload("Posts").Preload("Profile").Find(&users).Error

// BAD - N+1 query problem
users, _ := r.db.Find(&users)
for _, user := range users {
    posts, _ := getPosts(user.ID)  // Separate query for each user
}
```

**Select specific columns:**
```go
// GOOD - Select only needed columns
var users []User
r.db.Select("id, username, email").Find(&users)

// BAD - Select all columns
var users []User
r.db.Find(&users)  // Fetches unnecessary data
```

**Use indexes for common queries:**
```go
// Add indexes in migration
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

---

## Connection Management

**Set connection pool parameters:**
```go
// GOOD - Configure pool
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)      // Maximum idle connections
sqlDB.SetMaxOpenConns(100)     // Maximum open connections
sqlDB.SetConnMaxLifetime(time.Hour)  // Connection lifetime
```

**Close connections gracefully:**
```go
// GOOD - Cleanup function
func NewDB(dsn string) (*gorm.DB, func(), error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, nil, err
    }
    
    sqlDB, _ := db.DB()
    cleanup := func() {
        sqlDB.Close()
    }
    
    return db, cleanup, nil
}
```

---

## Error Handling

**Check for specific GORM errors:**
```go
// GOOD - Check for record not found
err := r.db.Where("username = ?", username).First(&user).Error
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("failed to find user: %w", err)
}

// GOOD - Wrap with context
if err := r.db.Create(&user).Error; err != nil {
    return fmt.Errorf("failed to create user %s: %w", user.Username, err)
}
```

---

## Soft Deletes

**Use soft deletes for data retention:**
```go
// GOOD - Soft delete
type User struct {
    gorm.Model
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Username  string
    Email     string
}

// Queries automatically exclude deleted records
r.db.Where("username = ?", username).First(&user)

// BAD - Hard delete
r.db.Unscoped().Delete(&user)  // Permanent deletion
```

**Include soft delete in queries:**
```go
// GOOD - Automatic soft delete handling
r.db.Find(&users)  // Excludes deleted records

// Include deleted records if needed
r.db.Unscoped().Find(&users)
```

---

## Hooks Usage

**Use hooks for data consistency:**
```go
// GOOD - BeforeCreate hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Email == "" {
        return errors.New("email required")
    }
    u.Email = strings.ToLower(u.Email)
    return nil
}

// GOOD - AfterUpdate hook
func (u *User) AfterUpdate(tx *gorm.DB) error {
    if u.Changed() {
        tx.Model(&u).Update("updated_at", time.Now())
    }
    return nil
}
```

---

## Common Pitfalls

**❌ Avoid these:**

```go
// BAD - Not using context
r.db.Find(&users)

// BAD - Not handling errors
r.db.Create(&user)  // Error ignored

// BAD - SQL injection risk
username := req.Username
r.db.Where(fmt.Sprintf("username = '%s'", username)).First(&user)

// BAD - N+1 queries
users, _ := r.db.Find(&users)
for _, user := range users {
    r.db.Where("user_id = ?", user.ID).Find(&posts)
}

// BAD - Not using transactions
r.db.Create(&order)
r.db.Create(&payment)  // If this fails, order orphaned
```

---

## Performance Best Practices

**Batch operations:**
```go
// GOOD - Batch insert
users := []User{user1, user2, user3}
r.db.CreateInBatches(users, 100)  // Batch size 100

// BAD - Individual inserts
for _, user := range users {
    r.db.Create(&user)  // N round trips
}
```

**Use indexes:**
```go
// GOOD - Query uses index
r.db.Where("username = ?", username).First(&user)  // Uses idx_users_username

// BAD - Query doesn't use index
r.db.Where("email LIKE ?", "%@example.com").Find(&users)  // Full table scan
```

**Limit results:**
```go
// GOOD - Limit results
r.db.Limit(100).Find(&users)

// GOOD - Pagination
r.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)
```

---

## Validation

**Use go-playground/validator/v10 for struct validation:**
```go
// GOOD - Use existing validator package
import validatePkg "github.com/go-playground/validator/v10"

type User struct {
    Username string `gorm:"uniqueIndex;not null;size:50" validate:"required,min=3,max=50,alphanum"`
    Email    string `gorm:"uniqueIndex;not null;size:100" validate:"required,email,max=100"`
    Password string `gorm:"not null;size:255" validate:"required,min=8"`
}
```

**CustomValidator pattern (matches existing implementation):**
```go
// GOOD - CustomValidator from internal/validator
type CustomValidator struct {
    validator *validatePkg.Validate
}

func NewCustomValidator() *CustomValidator {
    validate := validatePkg.New(validatePkg.WithRequiredStructEnabled())
    
    // Map JSON tags to validation
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
    
    return &CustomValidator{validator: validate}
}

func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        // Return detailed validation error
        for _, fieldErr := range err.(validatePkg.ValidationErrors) {
            property := FormatStructField(fieldErr)
            return &ValidationError{
                Property: property,
                Message:  "Validation failed for field: " + property,
            }
        }
    }
    return nil
}
```

**Common validation tags:**
```go
// Validation tag examples
Field string `validate:"required"`                    // Required
Field string `validate:"min=3,max=50"`               // Length
Field string `validate:"email"`                       // Email format
Field int    `validate:"min=1,max=100"`                // Range
Field string `validate:"alphanum"`                    // Alphanumeric
Field string `validate:"oneof=admin user guest"`      // Enum
Field string `validate:"required,unique"`              // Multiple
```

**Validation in repository/service:**
```go
// GOOD - Validate in service before database
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // 1. Struct validation
    if err := s.validator.Validate(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. Business validation (uniqueness)
    user := &User{Username: req.Username, Email: req.Email}
    if err := s.repo.ValidateUniqueness(ctx, user); err != nil {
        return err
    }
    
    // 3. Database operation
    return s.repo.CreateUser(ctx, user)
}
```

**Validation layers:**
```go
// GOOD - Multi-layer validation
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // Layer 1: Format validation (struct tags)
    if err := s.validator.Validate(req); err != nil {
        return err
    }
    
    // Layer 2: Business validation (uniqueness, rules)
    if err := s.ValidateBusinessRules(ctx, req); err != nil {
        return err
    }
    
    // Layer 3: Database operation
    return s.repo.CreateUser(ctx, req)
}
```

---

## Logging

**Log database operations appropriately:**
```go
// GOOD - Log errors with context
if err := r.db.Create(&user).Error; err != nil {
    r.log.Errorf("failed to create user %s: %v", user.Username, err)
    return err
}

// GOOD - Log slow queries
start := time.Now()
result := r.db.Where("username = ?", username).First(&user)
duration := time.Since(start)
if duration > 100*time.Millisecond {
    r.log.Warnf("slow query: %s took %v", "FindUserByUsername", duration)
}
```

---

## Testing

**Use test database for tests:**
```go
// GOOD - SQLite for tests
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    
    db.AutoMigrate(&User{})
    return db
}

// GOOD - Cleanup after tests
func teardownTestDB(t *testing.T, db *gorm.DB) {
    sqlDB, _ := db.DB()
    sqlDB.Close()
}
```

---

## Security

**Prevent SQL injection:**
```go
// GOOD - Parameterized queries
r.db.Where("username = ?", username).First(&user)

// BAD - String concatenation
r.db.Where("username = '" + username + "'").First(&user)  // SQL injection risk

// BAD - fmt.Sprintf
r.db.Where(fmt.Sprintf("username = '%s'", username)).First(&user)
```

**Whitelist columns:**
```go
// GOOD - Column whitelist
allowedColumns := map[string]bool{
    "username": true,
    "email":    true,
}

if !allowedColumns[field] {
    return fmt.Errorf("invalid field: %s", field)
}
query := fmt.Sprintf("%s = ?", field)
r.db.Where(query, value).First(&user)
```

---

## Best Practices Summary

| Practice | Do | Don't |
|----------|-----|-------|
| **Context** | Always pass `ctx` | Use `db` without context |
| **Transactions** | Use for multi-step operations | Skip transactions |
| **Preloading** | Use `Preload()` for associations | Query in loops (N+1) |
| **Selection** | `Select()` specific columns | Select all columns |
| **Errors** | Check `gorm.ErrRecordNotFound` | Ignore errors |
| **Soft Deletes** | Use `gorm.DeletedAt` | Hard delete by default |
| **Validation** | Validate before save | Save invalid data |
| **Security** | Parameterized queries | String concatenation |

---

## Database Field Conventions

**Use consistent naming and types:**
```go
// GOOD - Consistent field naming
type User struct {
    ID        int64     `gorm:"primarykey"`
    Username  string    `gorm:"uniqueIndex;not null;size:50"`
    Email     string    `gorm:"uniqueIndex;not null;size:100"`
    Password  string    `gorm:"not null" json:"-"`  // Never expose password
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// GOOD - Foreign key naming
type Order struct {
    ID         int64  `gorm:"primarykey"`
    UserID     int64  `gorm:"not null;index"`
    User       User   `gorm:"foreignKey:UserID"`
    TotalAmount float64 `gorm:"not null;default:0"`
    CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// BAD - Inconsistent naming
type User struct {
    id       int64
    userName string
    email    string
}
```

**Field tags:**
```go
// Common GORM tags
ID        int64     `gorm:"primarykey"`
Field     string    `gorm:"uniqueIndex"`
Field     string    `gorm:"index"`
Field     string    `gorm:"not null"`
Field     string    `gorm:"size:100"`
Field     int       `gorm:"default:0"`
CreatedAt time.Time `gorm:"autoCreateTime"`
UpdatedAt time.Time `gorm:"autoUpdateTime"`
DeletedAt gorm.DeletedAt `gorm:"index"`

// JSON tags for API
Password string `gorm:"not null" json:"-"`  // Never expose
Email    string `gorm:"uniqueIndex" json:"email"`
```

**Column naming:**
```go
// GOOD - Use struct field names (GORM snake_case conversion)
type UserProfile struct {
    ID           int64     `gorm:"primarykey"`
    FirstName    string    `gorm:"size:50;not null"`
    LastName     string    `gorm:"size:50;not null"`
    DateOfBirth  time.Time `gorm:"index"`
}

// Database columns: id, first_name, last_name, date_of_birth
```

---

## Validation Rules

**Validate before database operations:**
```go
// GOOD - Struct validation
type User struct {
    Username string `gorm:"uniqueIndex;not null;size:50" validate:"required,min=3,max=50,alphanum"`
    Email    string `gorm:"uniqueIndex;not null;size:100" validate:"required,email,max=100"`
    Password string `gorm:"not null;size:255" validate:"required,min=8"`
}

func (u *User) Validate() error {
    if err := validator.New().Struct(u); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// GOOD - Custom validation
func (u *User) Validate() error {
    if u.Username == "" {
        return ErrUsernameRequired
    }
    if len(u.Username) < 3 || len(u.Username) > 50 {
        return ErrInvalidUsernameLength
    }
    if !isValidEmail(u.Email) {
        return ErrInvalidEmail
    }
    if len(u.Password) < 8 {
        return ErrPasswordTooShort
    }
    return nil
}

// GOOD - Business logic validation
func (u *User) ValidateUnique(ctx context.Context, repo Repository) error {
    exists, err := repo.IsExistByUsername(ctx, u.Username)
    if err != nil {
        return err
    }
    if exists {
        return ErrUsernameAlreadyExists
    }
    return nil
}
```

**Use GORM callbacks for validation:**
```go
// GOOD - BeforeCreate hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Username == "" {
        return errors.New("username required")
    }
    u.Email = strings.ToLower(strings.TrimSpace(u.Email))
    return nil
}

// GOOD - BeforeUpdate hook
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    if u.Changed() {
        u.UpdatedAt = time.Now()
    }
    return u
}
```

**Validation patterns:**
```go
// GOOD - Validate before save
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    if err := user.Validate(); err != nil {
        return fmt.Errorf("invalid user: %w", err)
    }
    
    if err := r.checkUniqueness(ctx, user); err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Create(user).Error
}

// GOOD - Validate with context
func (r *Repository) checkUniqueness(ctx context.Context, user *User) error {
    usernameExists, emailExists, err := r.CheckUniqueness(ctx, user.Username, user.Email)
    if err != nil {
        return err
    }
    if usernameExists {
        return ErrUsernameAlreadyExists
    }
    if emailExists {
        return ErrEmailAlreadyExists
    }
    return nil
}
```

---

## Retry Patterns

**Retry transient errors:**
```go
// GOOD - Retry with exponential backoff
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    var err error
    for i := 0; i < 3; i++ {
        err = r.db.WithContext(ctx).Create(user).Error
        if err == nil {
            return nil
        }
        
        // Don't retry on validation errors
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return err
        }
        
        // Retry on connection errors
        if isTransientError(err) {
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }
        
        return err
    }
    return err
}

// GOOD - Retry helper
func Retry(fn func() error, maxAttempts int) error {
    var err error
    for i := 0; i < maxAttempts; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        if isTransientError(err) {
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }
        
        return err
    }
    return err
}

// Don't retry on these errors
func isTransientError(err error) bool {
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return false
    }
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return false
    }
    // Retry on connection timeouts, deadlocks
    return true
}
```

**Retry with context cancellation:**
```go
// GOOD - Retry with context
func (r *Repository) FindUser(ctx context.Context, id int64) (*User, error) {
    var user User
    var err error
    
    for i := 0; i < 3; i++ {
        err = r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
        if err == nil {
            return &user, nil
        }
        
        // Check context cancellation
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }
        
        if isTransientError(err) {
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }
        
        return nil, err
    }
    
    return nil, err
}
```

**Use retry libraries:**
```go
// GOOD - Use govern/http/retry or similar
import "github.com/haipham22/govern/http/retry"

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    return retry.Do(func() error {
        return r.db.WithContext(ctx).Create(user).Error
    },
        retry.WithMaxAttempts(3),
        retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)),
        retry.WithRetryIf(isTransientError),
    )
}
```

**Don't retry on validation errors:**
```go
// BAD - Retrying validation errors
for i := 0; i < 3; i++ {
    err := db.Create(&user).Error
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        time.Sleep(100 * time.Millisecond)
        continue  // ❌ Will never succeed
    }
}

// GOOD - Don't retry on errors that won't succeed
if errors.Is(err, gorm.ErrDuplicatedKey) {
    return ErrUserAlreadyExists  // ❌ Don't retry
}
```

---

## Connection Pool Management

**Configure connection pool properly:**
```go
// GOOD - Set pool parameters
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)           // Idle connections
sqlDB.SetMaxOpenConns(100)           // Max open connections  
sqlDB.SetConnMaxLifetime(time.Hour)  // Connection lifetime
sqlDB.SetConnMaxIdleTime(10 * time.Minute)  // Idle timeout
```

**Monitor pool health:**
```go
// GOOD - Check pool stats
func (r *Repository) PoolStats() PoolStats {
    sqlDB, _ := r.db.DB()
    stats := sqlDB.Stats()
    
    return PoolStats{
        MaxOpenConnections: stats.MaxOpenConnections,
        OpenConnections:    stats.OpenConnections,
        InUse:               stats.InUse,
        Idle:                stats.Idle,
        WaitCount:           stats.WaitCount,
        WaitDuration:        stats.WaitDuration,
    }
}
```

---

## Database Migration Rules

**Version control your migrations:**
```go
// GOOD - Use migration files
// migrations/20240628_create_users_table.up.sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

// migrations/20240628_create_users_table.down.sql
DROP INDEX idx_users_deleted_at;
DROP INDEX idx_users_email;
DROP INDEX idx_users_username;
DROP TABLE users;
```

**Use gorm:auto migrate for development only:**
```go
// GOOD - AutoMigrate in development
if os.Getenv("APP_ENV") == "development" {
    if err := db.AutoMigrate(&User{}, &Profile{}); err != nil {
        log.Fatalf("AutoMigration failed: %v", err)
    }
}

// PRODUCTION - Use manual migrations
// Don't use AutoMigrate in production
```

---

## Quick Reference

```go
// Context
db.WithContext(ctx).Where("id = ?", id).First(&user)

// Transaction
db.Transaction(func(tx *gorm.DB) error {
    return tx.Create(&user).Error
})

// Preload
db.Preload("Posts").Find(&users)

// Select specific columns
db.Select("id, username").Find(&users)

// Check record not found
err := db.First(&user).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, ErrNotFound
}

// Soft delete
db.Delete(&user)  // Sets deleted_at
db.Unscoped().Delete(&user)  // Hard delete
```
