# Clean Architecture Rules (bxcodec/go-clean-arch pattern)

## Folder Structure (Critical - Must Follow)

Based on bxcodec/go-clean-arch + Clean Architecture principles.

### domain/ (INNER LAYER - Enterprise Business Rules)

**Location:** `internal/domain/`

**Contains:**
- Pure business entities (structs with business logic methods)
- Domain-specific errors
- Value objects

**MUST NOT:**
- ❌ Define interfaces (interfaces defined by consuming layer)
- ❌ Import external packages (no frameworks, no databases)
- ❌ Depend on usecase/, repository/, handler/

**Rules:**
- Flat structure (NO subdirectories like domain/model/)
- Files use snake_case: user.go, product.go, errors.go
- Pure Go - no framework dependencies

**Example:**
```go
// internal/domain/user.go
package domain

type User struct {
    ID       int64
    Name     string
    Email    string
}

func (u *User) Validate() error {
    // Business validation logic
    if u.Email == "" {
        return ErrInvalidEmail
    }
    return nil
}
```

---

### usecase/ (MIDDLE LAYER - Application Business Rules)

**Location:** `internal/usecase/`

**Contains:**
- Use case implementations
- Repository interfaces (DEFINED by usecase, NOT domain)
- DTOs (Data Transfer Objects)
- Mocks for testing

**MUST NOT:**
- ❌ Depend on handler/, repository/ implementations
- ❌ Know about frameworks (Echo, gRPC, GORM)

**Rules:**
- Organized by business capability: usecase/auth/, usecase/product/, usecase/user/
- Each usecase folder: service.go (interface + impl), impl.go (use cases), dto.go, mocks/
- Interfaces defined HERE (service.go), not in domain/

**Example:**
```go
// internal/usecase/auth/service.go
package auth

// AuthRepository interface defined by auth usecase
type AuthRepository interface {
    FindByEmail(ctx context.Context, email string) (User, error)
    Create(ctx context.Context, user *User) error
}

type Service struct {
    authRepo AuthRepository
    logger   Logger
}

func NewService(repo AuthRepository, log Logger) *Service {
    return &Service{
        authRepo: repo,
        logger:   log,
    }
}

func (s *Service) RegisterUser(ctx context.Context, email, password string) error {
    // Use case logic
    existing, _ := s.authRepo.FindByEmail(ctx, email)
    if existing != (User{}) {
        return ErrUserAlreadyExists
    }
    
    user := &User{
        Email: email,
        Password: hashPassword(password),
    }
    
    return s.authRepo.Create(ctx, user)
}
```

---

### repository/ (OUTER LAYER - Frameworks & Drivers)

**Location:** `internal/repository/`

**Contains:**
- Database implementations (PostgreSQL, MongoDB, Redis)
- External service clients (Kafka, S3, HTTP clients)
- Infrastructure code

**MUST:**
- ✅ Implement interfaces from usecase/
- ✅ Use frameworks (GORM, Redis client, etc.)

**MUST NOT:**
- ❌ Define business logic
- ❌ Be imported by domain/

**Rules:**
- Nested by technology: repository/postgres/, repository/redis/, repository/kafka/
- Each implements interfaces from usecase/

**Example:**
```go
// internal/repository/postgres/auth.go
package postgres

import (
    "context"
    
    "path/to/usecase/auth"
    "gorm.io/gorm"
)

type authRepository struct {
    db *gorm.DB
}

// Implements auth.AuthRepository interface
func NewAuthRepository(db *gorm.DB) auth.AuthRepository {
    return &authRepository{db: db}
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (auth.User, error) {
    var user auth.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    return user, err
}

func (r *authRepository) Create(ctx context.Context, user *auth.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

---

### handler/ (OUTER LAYER - Delivery Mechanisms)

**Location:** `internal/handler/`

**Contains:**
- HTTP handlers (Echo framework)
- gRPC handlers
- Job workers
- Kafka consumers
- Middleware

**MUST:**
- ✅ Use usecase/ services
- ✅ Handle framework-specific concerns (binding, validation, routing)

**MUST NOT:**
- ❌ Contain business logic
- ❌ Access database directly (use repository/)

**Rules:**
- Organized by transport: handler/rest/, handler/grpc/, handler/job/, handler/kafka/
- Thin layer - delegate to usecase/ immediately

**Example:**
```go
// internal/handler/rest/auth.go
package rest

import (
    "net/http"
    
    "path/to/usecase/auth"
    "github.com/labstack/echo/v4"
)

type authHandler struct {
    authService *auth.Service
}

func NewAuthHandler(service *auth.Service) *authHandler {
    return &authHandler{authService: service}
}

func (h *authHandler) Login(c echo.Context) error {
    var req auth.LoginRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request"})
    }
    
    // Delegate to usecase
    token, err := h.authService.Login(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
    }
    
    return c.JSON(http.StatusOK, auth.LoginResponse{Token: token})
}
```

---

### bootstrap/ (Manual DI - Replacing Wire)

**Location:** `internal/bootstrap/`

**Contains:**
- Dependency injection constructors
- Setup functions (logger, database, HTTP server)

**MUST:**
- ✅ Wire all dependencies explicitly
- ✅ Return cleanup functions for resources

**Rules:**
- app.go: Main constructor
- database.go: Database setup with cleanup
- http.go: HTTP server setup
- logger.go: Logger setup

**Example:**
```go
// internal/bootstrap/app.go
package bootstrap

import (
    "path/to/internal/usecase/auth"
    "path/to/internal/repository/postgres"
    "path/to/internal/handler/rest"
)

func NewApp(cfg *Config) (*App, func(), error) {
    // 1. Logger
    logger := NewLogger(cfg.LogLevel)
    
    // 2. Database with cleanup
    db, cleanup, err := NewDatabase(cfg.Database)
    if err != nil {
        return nil, nil, err
    }
    
    // 3. Repositories (implement interfaces defined by usecases)
    authRepo := postgres.NewAuthRepository(db)
    productRepo := postgres.NewProductRepository(db)
    
    // 4. Use cases (depend on repository interfaces)
    authService := auth.NewService(authRepo, logger)
    productService := product.NewService(productRepo, logger)
    
    // 5. Handlers (depend on usecase services)
    authHandler := rest.NewAuthHandler(authService)
    productHandler := rest.NewProductHandler(productService)
    
    // 6. HTTP Server
    server := rest.NewServer(cfg.HTTP, authHandler, productHandler)
    
    return &App{
        Server: server,
        Logger: logger,
    }, cleanup, nil
}
```

---

## Dependency Rules (CRITICAL)

**Allowed dependencies (arrow means "can import"):**

```
handler/     → usecase/
usecase/      → domain/
repository/   → domain/ + usecase/ interfaces
bootstrap/    → everything
domain/       → NOTHING (no external dependencies)
```

**Forbidden dependencies:**

```
domain/    ✗ usecase/
domain/    ✗ repository/
domain/    ✗ handler/
usecase/   ✗ handler/
usecase/   ✗ repository/ implementations
handler/   ✗ repository/
```

---

## Interface Placement Rule (bxcodec pattern)

**Interfaces defined by CONSUMING layer:**

```
usecase/auth/service.go    → Defines AuthRepository interface
repository/postgres/       → Implements AuthRepository interface
```

**NOT:**

```
domain/repository.go       → ❌ WRONG: Don't define interfaces in domain
```

**Rationale:**
- Interface defined by who USES it (Dependency Inversion Principle)
- Domain doesn't know about repositories
- Use case defines what it needs

---

## Complete Folder Structure

```
internal/
├── bootstrap/                    # Manual DI wiring
│   ├── app.go
│   ├── logger.go
│   ├── database.go
│   ├── http.go
│   └── worker.go
├── usecase/                      # Use cases (MIDDLE LAYER)
│   ├── auth/
│   │   ├── service.go            # AuthRepository interface
│   │   ├── impl.go
│   │   ├── dto.go
│   │   └── mocks/
│   ├── product/
│   └── user/
├── domain/                       # Entities (INNER LAYER)
│   ├── user.go
│   ├── product.go
│   └── errors.go
├── repository/                   # Implementations (OUTER LAYER)
│   ├── helper.go
│   ├── postgres/
│   ├── redis/
│   └── kafka/
├── handler/                      # Delivery (OUTER LAYER)
│   ├── rest/
│   ├── grpc/
│   ├── job/
│   └── kafka/
└── errors/                       # Custom errors
```

---

