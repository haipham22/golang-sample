# Database Rules

**Rules for GORM ORM and PostgreSQL database operations.**

---

## GORM (Database ORM)

**Repository structure with error envelope:**
```go
// GOOD - Proper GORM repository with error envelope pattern
type postgresUserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &postgresUserRepository{db: db}
}

// Error envelope for database operations
type DBError struct {
    Op       string // Operation that failed
    Table    string // Table being accessed
    Err      error  // Underlying error
    Severity string // "transient" or "permanent"
}

func (e *DBError) Error() string {
    return fmt.Sprintf("%s on %s: %v", e.Op, e.Table, e.Err)
}

func (e *DBError) Unwrap() error {
    return e.Err
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id int64) (User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Where("id = ? AND deleted_at IS NULL", id).
        First(&user).Error
        
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return User{}, &errors.AppError{
                Code: errors.ErrCodeNotFound,
                Err: &DBError{
                    Op:    "find_by_id",
                    Table: "users",
                    Err:   err,
                },
                Message: fmt.Sprintf("user with ID %d not found", id),
            }
        }
        return User{}, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "find_by_id",
                Table:    "users",
                Err:      err,
                Severity: "transient",
            },
            Message: "database error",
        }
    }
    
    return user, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, user *User) error {
    err := r.db.WithContext(ctx).Create(user).Error
    if err != nil {
        return &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "create",
                Table:    "users",
                Err:      err,
                Severity: classifyErrorSeverity(err),
            },
            Message: "failed to create user",
        }
    }
    return nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *User) error {
    result := r.db.WithContext(ctx).
        Model(user).
        Updates(map[string]interface{}{
            "name":      user.Name,
            "email":     user.Email,
            "updated_at": time.Now(),
        })
        
    if result.Error != nil {
        return &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "update",
                Table:    "users",
                Err:      result.Error,
                Severity: classifyErrorSeverity(result.Error),
            },
            Message: "failed to update user",
        }
    }
    
    if result.RowsAffected == 0 {
        return &errors.AppError{
            Code: errors.ErrCodeNotFound,
            Err: &DBError{
                Op:    "update",
                Table: "users",
                Err:   gorm.ErrRecordNotFound,
            },
            Message: fmt.Sprintf("user with ID %d not found", user.ID),
        }
    }
    
    return nil
}

// Helper function to classify error severity
func classifyErrorSeverity(err error) string {
    // Connection errors, timeouts -> transient
    if errors.Is(err, context.DeadlineExceeded) || 
       errors.Is(err, context.Canceled) ||
       isConnectionError(err) {
        return "transient"
    }
    // Constraint violations, data errors -> permanent
    return "permanent"
}

func isConnectionError(err error) bool {
    // Check for connection-related errors
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "broken pipe")
}
```

**GORM rules:**
- ✅ ALWAYS use `WithContext(ctx)` for queries
- ✅ Wrap GORM errors with error envelope pattern
- ✅ Use transactions for multi-step operations
- ✅ Check `RowsAffected` for updates
- ✅ Use soft deletes (`deleted_at`)
- ✅ Classify error severity (transient vs permanent)
- ❌ NEVER ignore query errors
- ❌ NEVER use `*gorm.DB` in domain layer
- ❌ NEVER return raw GORM errors to upper layers

**Transaction handling:**
```go
func (r *postgresOrderRepository) CreateOrderWithItems(ctx context.Context, order *Order, items []OrderItem) error {
    err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Create order
        if err := tx.Create(order).Error; err != nil {
            return &DBError{
                Op:    "create_order",
                Table: "orders",
                Err:   err,
            }
        }
        
        // Create items
        for i := range items {
            items[i].OrderID = order.ID
            if err := tx.Create(&items[i]).Error; err != nil {
                return &DBError{
                    Op:    "create_order_item",
                    Table: "order_items",
                    Err:   err,
                }
            }
        }
        
        // Update inventory
        for _, item := range items {
            result := tx.Exec(
                "UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?",
                item.Quantity, item.ProductID, item.Quantity,
            )
            if result.Error != nil {
                return &DBError{
                    Op:    "update_inventory",
                    Table: "products",
                    Err:   result.Error,
                }
            }
            if result.RowsAffected == 0 {
                return &errors.AppError{
                    Code: errors.ErrCodeConflict,
                    Err: &DBError{
                        Op:    "update_inventory",
                        Table: "products",
                        Err:   errors.New("insufficient stock"),
                    },
                    Message: fmt.Sprintf("insufficient stock for product %d", item.ProductID),
                }
            }
        }
        
        return nil
    })
    
    if err != nil {
        // Wrap transaction error in envelope
        if dbErr, ok := err.(*DBError); ok {
            return &errors.AppError{
                Code: errors.ErrCodeInternal,
                Err:  dbErr,
                Message: "transaction failed",
            }
        }
        if appErr, ok := err.(*errors.AppError); ok {
            return appErr
        }
        return &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err:  err,
            Message: "transaction error",
        }
    }
    
    return nil
}
```

---

## PostgreSQL (Database)

**Connection setup with error envelope:**
```go
// GOOD - Proper connection setup with error envelope
func NewDatabase(cfg DatabaseConfig) (*gorm.DB, func(), error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })
    if err != nil {
        return nil, nil, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "connect",
                Table:    "",
                Err:      err,
                Severity: "permanent",
            },
            Message: "failed to connect to database",
        }
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, nil, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "get_instance",
                Table:    "",
                Err:      err,
                Severity: "permanent",
            },
            Message: "failed to get database instance",
        }
    }
    
    // Connection pool settings
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // Verify connection
    if err := sqlDB.Ping(); err != nil {
        return nil, nil, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "ping",
                Table:    "",
                Err:      err,
                Severity: "transient",
            },
            Message: "failed to ping database",
        }
    }
    
    // Cleanup function
    cleanup := func() {
        sqlDB.Close()
    }
    
    return db, cleanup, nil
}
```

**PostgreSQL rules:**
- ✅ Use connection pooling
- ✅ Set appropriate timeouts
- ✅ Handle connection errors with envelope
- ✅ Verify connection with Ping()
- ✅ Wrap connection errors with custom error types
- ✅ Classify error severity (transient vs permanent)
- ❌ NEVER use hardcoded credentials
- ❌ NEVER trust user input in queries (use parameterized)
- ❌ NEVER expose database errors directly to clients

**Migration with error envelope:**
```go
// GOOD - Migration with proper error handling
func Migrate(db *gorm.DB) error {
    // Auto-migrate (development only)
    if err := db.AutoMigrate(&User{}, &Product{}, &Order{}); err != nil {
        return &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &DBError{
                Op:       "migrate",
                Table:    "schema",
                Err:      err,
                Severity: "permanent",
            },
            Message: "failed to run migrations",
        }
    }
    
    return nil
}

// For production, use proper migration tools:
// - golang-migrate/migrate
// - pressly/goose
// - pressly/goose
```
