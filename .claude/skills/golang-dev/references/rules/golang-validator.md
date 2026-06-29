# Go Validation Rules

**Best practices for input validation using go-playground/validator and custom validation patterns.**

---

## Validator Setup

**Use go-playground/validator/v10:**
```go
import validatePkg "github.com/go-playground/validator/v10"

type CustomValidator struct {
    validator *validatePkg.Validate
}

func NewCustomValidator() *CustomValidator {
    validate := validatePkg.New(validatePkg.WithRequiredStructEnabled())
    
    // Map JSON tags to validation fields
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
    
    return &CustomValidator{validator: validate}
}
```

---

## Common Validation Tags

**Required & Length:**
```go
Field string `validate:"required"`              // Must be non-empty
Field string `validate:"min=3,max=50"`          // String length
Field int    `validate:"min=1,max=100"`           // Number range
Field int    `validate:"gte=0,lte=100"`          // Greater/less than
```

**Format:**
```go
Field string `validate:"email"`                  // Email format
Field string `validate:"url"`                    // URL format
Field string `validate:"uri"`                    // URI format
Field string `validate:"uuid"`                   // UUID format
Field string `validate:"hostname"`              // Hostname format
Field string `validate:"ip"`                     // IP address
Field string `validate:"ipv4"`                   // IPv4 address
Field string `validate:"ipv6"`                   // IPv6 address
```

**String Patterns:**
```go
Field string `validate:"alphanum"`              // Alphanumeric only
Field string `validate:"alpha"`                  // Letters only
Field string `validate:"numeric"`                // Numbers only
Field string `validate:"ascii"`                   // ASCII only
Field string `validate:"lowercase"`               // Lowercase only
Field string `validate:"uppercase"`               // Uppercase only
Field string `validate:"contains=@example"`      // Must contain substring
Field string `validate:"startswith=admin"`        // Must start with prefix
Field string `validate:"endswith=.com"`          // Must end with suffix
```

**Comparison:**
```go
Field int `validate:"eq=10"`                   // Equal to
Field int `validate:"ne=10"`                   // Not equal to
Field int `validate:"gt=10"`                   // Greater than
Field int `validate:"gte=10"`                  // Greater or equal
Field int `validate:"lt=100"`                  // Less than
Field int `validate:"lte=100"`                 // Less or equal
```

**Enums:**
```go
Field string `validate:"oneof=admin user guest"`  // Must be in list
Field int    `validate:"oneof=1 2 3"`               // Must be in list
```

**Date/Time:**
```go
Field time.Time `validate:"required"`           // Required timestamp
Field time.Time `validate:"gt"`                  // Must be after current time
Field time.Time `validate:"lte=2099-12-31"`     // Before specific date
```

**Multiple Tags:**
```go
Field string `validate:"required,min=3,max=50,alphanum"`
```

---

## Struct Level Validation

**Enable struct level validation:**
```go
validate := validatePkg.New(validatePkg.WithRequiredStructEnabled())
```

**Use dive modifier for nested structs:**
```go
type Request struct {
    User User `validate:"dive"`
}

type User struct {
    Name  string `validate:"required"`
    Email string `validate:"required,email"`
}
```

---

## Custom Validation Functions

**Register custom validators:**
```go
func validateFlUsername(fl validator.FieldLevel) bool {
    field := fl.Field().String()
    // Custom validation logic
    return len(field) >= 3 && len(field) <= 50 && isAlphanumeric(field)
}

func validateUniqueEmail(fl validator.FieldLevel) bool {
    email := fl.Field().String()
    // Check database for uniqueness
    return !isEmailExists(email)
}

// Register custom validators
validate.RegisterValidation("flusername", validateFlUsername)
validate.RegisterValidation("uniqueemail", validateUniqueEmail)

// Use in struct
Username string `validate:"required,flusername"`
Email    string `validate:"required,uniqueemail"`
```

**Cross-field validation:**
```go
func validatePasswordMatch(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    confirm := fl.Top().Field().String()
    return password == confirm
}

// Register struct level validation
validate.RegisterStructValidation(PasswordMatchStructLevelValidation)

func PasswordMatchStructLevelValidation(sl validator.StructLevel) bool {
    user := sl.Current().Interface().(User)
    return user.Password == user.ConfirmPassword
}
```

---

## Error Handling

**Format validation errors:**
```go
func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        for _, fieldErr := range err.(validatePkg.ValidationErrors) {
            property := FormatStructField(fieldErr)
            detail := ErrorDetail{
                Property: property,
                Msg:      "Validation failed for field: " + property,
                Tag:      fieldErr.Tag(),
                Value:     fmt.Sprintf("%v", fieldErr.Value()),
            }
            return &ValidationError{Detail: detail}
        }
    }
    return nil
}
```

**Error types:**
```go
type ValidationError struct {
    Detail ErrorDetail
}

func (e *ValidationError) Error() string {
    return e.Detail.Msg
}

func (e *ValidationError) GetProperty() string {
    return e.Detail.Property
}
```

---

## Validation Patterns

**Service layer validation:**
```go
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // 1. Struct validation
    if err := s.validator.Validate(req); err != nil {
        return fmt.Errorf("invalid request: %w", err)
    }
    
    // 2. Business validation
    if err := s.ValidateBusinessRules(ctx, req); err != nil {
        return err
    }
    
    // 3. Database operation
    return s.repo.CreateUser(ctx, req)
}
```

**Middleware validation:**
```go
func ValidateRequest(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        var req CreateUserRequest
        if err := c.Bind(&req); err != nil {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "invalid request format",
            })
        }
        
        if err := validator.Validate(req); err != nil {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "validation failed",
                Details: err,
            })
        }
        
        return next(c)
    }
}
```

**Repository validation:**
```go
func (r *Repository) ValidateUniqueness(ctx context.Context, username, email string) error {
    // Check database for uniqueness
    var count int64
    r.db.WithContext(ctx).Model(&User{}).Where("username = ?", username).Count(&count)
    if count > 0 {
        return ErrUsernameAlreadyExists
    }
    
    r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count)
    if count > 0 {
        return ErrEmailAlreadyExists
    }
    
    return nil
}
```

---

## Validation Layers

**Three-layer validation approach:**

| Layer | Type | Purpose | Example |
|-------|------|---------|---------|
| **1. Request** | Format | Validate HTTP input | `validate:"required,email"` |
| **2. Business** | Logic | Validate against rules | Uniqueness, permissions |
| **3. Database** | Constraints | Validate at DB level | `uniqueIndex`, `not null` |

**Validation flow:**
```go
// GOOD - Three-layer validation
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // Layer 1: Request format validation
    if err := s.validator.Validate(req); err != nil {
        return fmt.Errorf("invalid request: %w", err)
    }
    
    // Layer 2: Business logic validation
    if err := s.ValidateBusinessRules(ctx, req); err != nil {
        return err
    }
    
    // Layer 3: Database constraint validation
    return s.repo.CreateUser(ctx, req)  // Will fail if constraint violated
}
```

---

## Best Practices

**✅ DO:**
- Validate at service layer before database
- Use struct tags for common validations
- Create custom validators for business logic
- Return detailed validation errors
- Use three-layer validation approach

**❌ DON'T:**
- Skip validation before database operations
- Mix validation with business logic
- Return generic validation messages
- Trust client input without validation
- Validate only at database layer

---

## Quick Reference

**Common Tags:**
```go
`validate:"required"`              // Required
`validate:"min=3,max=50"`          // Length
`validate:"email"`                  // Email format
`validate:"alphanum"`              // Alphanumeric
`validate:"oneof=a b c"`            // Enum
`validate:"gte=18"`                 // Minimum age
`validate:"required,unique"`       // Multiple tags
```

**Setup:**
```go
validator := validatePkg.New(validatePkg.WithRequiredStructEnabled())
validator.RegisterTagNameFunc(jsonTagFunc)
```

**Error Handling:**
```go
if err := validator.Validate(req); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
```
