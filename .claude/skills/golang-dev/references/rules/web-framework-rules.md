# Web Framework Rules

**Rules for Echo HTTP framework and request validation.**

---

## Echo Framework (HTTP)

**Handler structure with custom error types:**
```go
// GOOD - Proper Echo handler with custom error types
type authHandler struct {
    authService *auth.Service
}

func NewAuthHandler(service *auth.Service) *authHandler {
    return &authHandler{authService: service}
}

func (h *authHandler) Login(c echo.Context) error {
    var req auth.LoginRequest
    if err := c.Bind(&req); err != nil {
        return errors.InvalidInput("request", "Invalid request format")
    }
    
    // Delegate to usecase
    resp, err := h.authService.Login(c.Request().Context(), req)
    if err != nil {
        return err // Return custom error type directly
    }
    
    return c.JSON(http.StatusOK, SuccessResponse{Data: resp})
}
```

**Centralized error handler:**
```go
// GOOD - Centralized error handler with custom error types
func HTTPErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
    return func(err error, c echo.Context) {
        // Extract request ID from context
        requestID := c.Get("request_id")
        
        // Log error with context
        logger.Error("request error",
            zap.String("request_id", requestID),
            zap.String("path", c.Path()),
            zap.Error(err),
        )
        
        // Handle custom error types
        var appErr *errors.AppError
        if errors.As(err, &appErr) {
            // Map custom error code to HTTP status
            statusCode := appErr.Code.HTTPStatus()
            message := appErr.Message
            
            // Log internal errors differently
            if statusCode >= 500 {
                logger.Error("internal error",
                    zap.String("code", string(appErr.Code)),
                    zap.Error(appErr.Err),
                )
            }
            
            _ = c.JSON(statusCode, map[string]interface{}{
                "code":       string(appErr.Code),
                "message":    message,
                "request_id": requestID,
            })
            return
        }
        
        // Handle Echo HTTP errors
        if he, ok := err.(*echo.HTTPError); ok {
            code := he.Code
            message := fmt.Sprintf("%s", he.Message)
            
            if code >= 500 {
                logger.Error("http error", zap.Int("status", code), zap.String("message", message))
            } else {
                logger.Warn("http error", zap.Int("status", code), zap.String("message", message))
            }
            
            _ = c.JSON(code, map[string]interface{}{
                "code":       errors.CodeFor(code),
                "message":    message,
                "request_id": requestID,
            })
            return
        }
        
        // Fallback for unknown errors
        logger.Error("unknown error", zap.Error(err))
        _ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "code":       string(errors.ErrCodeInternal),
            "message":    "Internal server error",
            "request_id": requestID,
        })
    }
}
```

**Echo rules:**
- ✅ Use struct for handlers (dependency injection)
- ✅ Return `echo.Context` from handlers
- ✅ Use `c.Bind()` for request binding
- ✅ Use `c.Validate()` for validation
- ✅ Return custom error types directly
- ✅ Centralized error handler with custom errors
- ❌ NEVER use switch/case for error mapping
- ❌ NEVER ignore context cancellation

**Middleware usage:**
```go
// GOOD - Proper middleware with custom errors
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        token := c.Request().Header.Get("Authorization")
        if token == "" {
            return errors.Unauthorized("missing_token", "Authorization token required")
        }
        
        claims, err := validateToken(token)
        if err != nil {
            return errors.Unauthorized("invalid_token", "Invalid token")
        }
        
        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("request_id", generateRequestID())
        
        return next(c)
    }
}

// Register middleware
e.Use(middleware.RequestID())
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORS())
api := e.Group("/api")
api.Use(AuthMiddleware)
```

---

## Validation

**Request validation with custom errors:**
```go
// GOOD - Proper validation with custom error types
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

func (h *authHandler) Register(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return errors.InvalidInput("request", "Invalid request format")
    }
    
    // Validate using echo's validator
    if err := c.Validate(req); err != nil {
        return errors.InvalidInput("validation", formatValidationError(err))
    }
    
    // Delegate to usecase
    user, err := h.authService.Register(c.Request().Context(), req)
    if err != nil {
        return err // Return custom error from service
    }
    
    return c.JSON(http.StatusCreated, SuccessResponse{Data: user})
}

// Format validation errors with custom error types
func formatValidationError(err error) string {
    var errs validator.ValidationErrors
    if errors.As(err, &errs) {
        var messages []string
        for _, e := range errs {
            switch e.Tag() {
            case "required":
                messages = append(messages, fmt.Sprintf("%s is required", e.Field()))
            case "email":
                messages = append(messages, fmt.Sprintf("%s must be a valid email", e.Field()))
            case "min":
                messages = append(messages, fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param()))
            case "max":
                messages = append(messages, fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param()))
            case "oneof":
                messages = append(messages, fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param()))
            default:
                messages = append(messages, fmt.Sprintf("%s is invalid (%s)", e.Field(), e.Tag()))
            }
        }
        return strings.Join(messages, "; ")
    }
    return err.Error()
}
```

**Validation rules:**
- ✅ Use struct tags for validation
- ✅ Validate in handler layer
- ✅ Return custom error types (InvalidInput)
- ✅ Provide clear, actionable error messages
- ✅ Validate business rules in usecase layer
- ✅ Use appropriate validation tags (email, min, max, required, oneof)
- ❌ NEVER trust client-side validation only
- ❌ NEVER expose internal validation errors directly to clients
- ❌ NEVER ignore validation errors

**Business validation in usecase:**
```go
// GOOD - Business logic validation in usecase
func (s *Service) RegisterUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Check if user already exists (business rule)
    existing, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, ErrUserNotFound) {
        return nil, err // Return unexpected errors
    }
    
    if existing.ID != 0 {
        return nil, errors.Conflict("user_exists", "user with this email already exists")
    }
    
    // Validate password strength (business rule)
    if !isStrongPassword(req.Password) {
        return nil, errors.InvalidInput("password", "password does not meet security requirements")
    }
    
    // Create user
    user := &User{
        Email:    req.Email,
        Password: hashPassword(req.Password),
        Name:     req.Name,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```
