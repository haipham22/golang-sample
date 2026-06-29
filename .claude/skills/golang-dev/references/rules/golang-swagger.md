# Go Swagger/OpenAPI Rules

**Best practices for API documentation using Swagger/OpenAPI and swaggo/swag.**

---

## Swagger Setup

**swag is managed by mise:**
```toml
# mise.toml
[tools]
swag = "latest"
```

**Install/update tools:**
```bash
mise install
```

**Use swag via mise:**
```bash
mise exec -- swag init -g cmd/api/main.go
mise exec -- swag generate
mise exec -- swag fmt
```

**Initialize Swagger documentation:**
```bash
swag init -g cmd/api/main.go
```

**Directory structure:**
```
docs/
├── swagger/                  # Swagger UI assets
├── docs.go                   # Generated swagger docs
└── swagger.yaml             # Swagger specification
```

---

## Swagger Annotations

**API Operation Annotations:**
```go
// @Summary Create new user
// @Description Create a new user account with email and password
// @Tags users
// @Accept  json
// @Produce  json
// @Param   request  body      CreateUserRequest  true  "Request body"
// @Success 200     {object}  User                  "User created successfully"
// @Failure 400     {object}  ErrorResponse          "Invalid request"
// @Failure 409     {object}  ErrorResponse          "User already exists"
// @Router  /api/users [post]
func (h *Handler) CreateUser(c echo.Context) error {
    // Handler implementation
}
```

**@Summary vs @Description:**
```go
// GOOD - Brief summary
// @Summary Create user
// @Description Create a new user with email verification

// BAD - Too long
// @Summary Create a new user account with automatic email verification sent to the provided email address
```

**@Tags for Grouping:**
```go
// @Tags users
// @Tags users,admin
// @Tags auth
```

**@Router for Route Definition:**
```go
// @Router /api/users [post]
// @Router /api/users/{id} [get]
// @Router /api/users/{id} [put]
// @Router /api/users/{id} [delete]
```

---

## Request/Response Annotations

**Request Parameters:**
```go
// Query parameter
// @Param   id    path      int  true  "User ID"
// @Param   name  query     string  false  "Filter by name"
// @Param   page  query     int     false  "Page number" default(1)

// Body parameter
// @Param   request  body      CreateUserRequest  true  "Request body"

// Header parameter
// @Param   Authorization  header  string  true  "Bearer token"
```

**Success Responses:**
```go
// @Success 200 {object} User "User created successfully"
// @Success 200 {array} User "List of users"
// @Success 204 "No content (deleted successfully)"
```

**Error Responses:**
```go
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
```

---

## Model Definitions

**Define request/response models:**
```go
// User response model
type UserResponse struct {
    ID       int64  `json:"id" example:"1"`
    Username string `json:"username" example:"john_doe"`
    Email    string `json:"email" example:"john@example.com"`
}

// Create user request
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// Error response
type ErrorResponse struct {
    Code    string `json:"code" example:"VALIDATION_ERROR"`
    Message string `json:"message" example:"Invalid request format"`
}
```

**Use models in annotations:**
```go
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
```

---

## Security Annotations

**Authentication:**
```go
// @Security BearerAuth
// @SecurityDefinitions BearerAuth
// @description Bearer Auth authentication
// @name Authorization
// @in header
// @Type http
// @scheme bearer
func (h *Handler) GetProfile(c echo.Context) error {
    // Protected handler
}
```

**Multiple security schemes:**
```go
// @Security BearerAuth
// @Security ApiKeyAuth
func (h *Handler) AdminAction(c echo.Context) error {
    // Supports both Bearer and API key
}
```

---

## Swagger Development Workflow

**Development workflow:**
```bash
# 1. Add annotations to handlers
# @Summary Create user
// @Router /api/users [post]

# 2. Generate swagger docs
mise exec -- swag fmt
mise exec -- swag generate

# 3. View Swagger UI
# http://localhost:8080/swagger/

# 4. Format generated code
mise exec -- goimports -w .
```

**Auto-generate on build:**
```go
//go:generate swag fmt
//go:generate swag init -g cmd/api/main.go
```

---

**CRUD Operations:**
```go
// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Request body"
// @Success 201 {object} UserResponse "User created"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "User exists"
// @Router /api/users [post]

// @Summary Get user by ID
// @Description Get user information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse "User found"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/users/{id} [get]

// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Request body"
// @Success 200 {object} UserResponse "User updated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/users/{id} [put]

// @Summary Delete user
// @Description Delete user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "User deleted"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /api/users/{id} [delete]
```

**List/Search Operations:**
```go
// @Summary List users
// @Description Get paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by username"
// @Success 200 {array} UserResponse "List of users"
// @Router /api/users [get]

// @Summary Search users
// @Description Search users by username or email
// @Tags users
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} UserResponse "Search results"
// @Router /api/users/search [get]
```

---

## Authentication Annotations

**JWT Authentication:**
```go
// @Summary Login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Router /api/auth/login [post]

// LoginResponse model
type LoginResponse struct {
    Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0"`
}
```

---

## Example Values

**Use example tags for clarity:**
```go
// @Param id path int true "User ID" minimum(1) example(1)
// @Param email query string false "Filter by email" example(user@example.com)
// @Success 200 {object} UserResponse "User found" example({"id": 1,"username":"john_doe"})
// @Failure 404 {object} ErrorResponse "User not found" example({"code":"NOT_FOUND","message":"User not found"})
```

---

## Common Pitfalls

**❌ Avoid these:**

```go
// BAD - Missing @Summary
// @Tags users
// @Router /api/users [get]

// BAD - Too long @Description
// @Description This endpoint retrieves user information from the database including all associated data like profiles, settings, preferences, and historical activity logs

// BAD - Missing @Param
// @Router /api/users/{id} [get]

// BAD - Wrong @Param type
// @Param id body int true "User ID"  // Should be path

// BAD - Missing @Router
// @Summary Get user
// @Tags users
```

---

## Best Practices

**✅ DO:**
- Add @Summary for all operations
- Use @Tags for logical grouping
- Specify all parameters with @Param
- Document all response codes
- Add example values for clarity
- Use @Security for protected endpoints
- Keep @Description concise

**❌ DON'T:**
- Skip @Router annotation
- Omit parameter documentation
- Forget error response codes
- Write verbose descriptions
- Mix operation types in one annotation
- Document internal implementation details

---

## Swagger Development Workflow

**Generate swagger docs:**
```bash
swag fmt
swag generate
```

**View Swagger UI:**
- Local: `http://localhost:8080/swagger/`
- JSON spec: `http://localhost:8080/swagger/doc.json`

**Auto-generate on build:**
```go
//go:generate swag fmt
//go:generate swag init -g cmd/api/main.go
```

---

## Quick Reference

**Basic CRUD Template:**
```go
// @Summary Create item
// @Tags items
// @Router /api/items [post]

// @Summary Get item by ID
// @Tags items
// @Router /api/items/{id} [get]

// @Summary Update item
// @Tags items
// @Router /api/items/{id} [put]

// @Summary Delete item
// @Tags items
// @Router /api/items/{id} [delete]

// @Summary List items
// @Tags items
// @Router /api/items [get]
```

**Annotation Order:**
1. @Summary
2. @Description (optional, concise)
3. @Tags
4. @Accept / @Produce
5. @Security (if protected)
6. @Param (for each parameter)
7. @Success (for each success code)
8. @Failure (for each error code)
9. @Router

---

## Security Considerations

**Disable Swagger in production:**
```go
if os.Getenv("APP_ENV") != "production" {
    e := echo.New()
    httpEcho.WithEchoSwagger(e,
        httpEcho.WithSwaggerEnabled(true),
    )
}
```

**Protect Swagger endpoint:**
```go
// @Security BearerAuth
func GetSwaggerDocs(c echo.Context) error {
    // Requires authentication
    return c.File("docs/swagger/swagger.yaml")
}
```

**Don't expose sensitive data:**
```go
// GOOD - Generic example
// @Success 200 {object} UserResponse "User found" example({"id": 1,"username":"john_doe"})

// BAD - Real sensitive data
// @Success 200 {object} UserResponse "User found" example({"id": 12345,"email":"secret@company.com","password":"hash123"})
```
