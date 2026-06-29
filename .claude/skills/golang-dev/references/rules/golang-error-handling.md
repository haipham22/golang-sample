# Go Error Handling Rules

**Best practices for error handling, wrapping, and custom errors in Go.**

---

## Error Wrapping

**Wrap errors with context:**
```go
// GOOD - Error wrapping with context
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil {
        return fmt.Errorf("failed to find user by email %s: %w", req.Email, err)
    }
    
    if user.ID != 0 {
        return fmt.Errorf("user already exists: %s", req.Email)
    }
    
    return s.repo.Create(ctx, &user)
}

// BAD - Lose error context
func (s *Service) CreateUser(req CreateUserRequest) error {
    user, err := s.repo.FindByEmail(req.Email)
    if err != nil {
        return err // Lost context about what operation failed
    }
    return nil
}
```

**Custom error types:**
```go
// Define error types for domain-specific errors
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrUserExists      = errors.New("user already exists")
    ErrInvalidEmail    = errors.New("invalid email format")
)

// Use in code
func (s *Service) GetUser(id int64) (User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return User{}, ErrUserNotFound
        }
        return User{}, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

---

## Error Checking Rules

**ALWAYS check errors:**
```go
// GOOD
result, err := someFunction()
if err != nil {
    // Handle error
}

// BAD - Never ignore errors
result, _ := someFunction()
```

**Use errors.Is() and errors.As():**
```go
// GOOD - Check error type
if errors.Is(err, ErrUserNotFound) {
    // Handle user not found
}

// GOOD - Extract error type
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", validationErr.Fields)
}
```

---

## Error Return Patterns

**Return early, return often:**
```go
// GOOD - Early returns
func (s *Service) ProcessUser(id int64) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return err
    }
    
    if !user.IsActive {
        return ErrUserInactive
    }
    
    if err := s.validateUser(user); err != nil {
        return err
    }
    
    return s.process(&user)
}

// BAD - Nested if statements
func (s *Service) ProcessUser(id int64) error {
    user, err := s.repo.FindByID(id)
    if err == nil {
        if user.IsActive {
            if err := s.validateUser(user); err == nil {
                return s.process(&user)
            }
            return err
        }
        return ErrUserInactive
    }
    return err
}
```

---

## Custom Error Design

**Structured errors:**
```go
// GOOD - Error with details
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}

// Use
func (v *Validator) ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return &ValidationError{
            Field:   "email",
            Message: "must contain @ symbol",
        }
    }
    return nil
}
```

**Sentinel errors:**
```go
// GOOD - Predefined errors for comparison
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized access")
    ErrValidation   = errors.New("validation failed")
)

// Check
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

---

## Error Handling in Different Layers

**Repository layer:**
```go
func (r *userRepository) FindByEmail(ctx context.Context, email string) (User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return User{}, ErrUserNotFound
        }
        return User{}, fmt.Errorf("failed to find user: %w", err)
    }
    return user, nil
}
```

**Service layer:**
```go
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return "", ErrInvalidCredentials
        }
        return "", fmt.Errorf("failed to find user: %w", err)
    }
    
    if !s.verifyPassword(password, user.Password) {
        return "", ErrInvalidCredentials
    }
    
    token, err := s.generateToken(user.ID)
    if err != nil {
        return "", fmt.Errorf("failed to generate token: %w", err)
    }
    
    return token, nil
}
```

**Handler layer:**
```go
func (h *authHandler) Login(c echo.Context) error {
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{
            Message: "invalid request format",
        })
    }
    
    token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
    if err != nil {
        if errors.Is(err, ErrInvalidCredentials) {
            return c.JSON(http.StatusUnauthorized, ErrorResponse{
                Message: "invalid credentials",
            })
        }
        return c.JSON(http.StatusInternalServerError, ErrorResponse{
            Message: "internal server error",
        })
    }
    
    return c.JSON(http.StatusOK, LoginResponse{Token: token})
}
```

---

## Logging Errors

**Structured logging with context:**
```go
// GOOD - Log with context
func (s *Service) ProcessUser(id int64) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        s.logger.Errorw("failed to find user",
            "user_id", id,
            "error", err,
        )
        return fmt.Errorf("failed to find user %d: %w", id, err)
    }
    
    return nil
}

// BAD - Log without context
func (s *Service) ProcessUser(id int64) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        s.logger.Error(err.Error()) // No context
        return err
    }
    
    return nil
}
```
