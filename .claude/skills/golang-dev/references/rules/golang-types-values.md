# Go Types and Values Rules

**Best practices for using pointers, values, generics, return values, and variables in Go.**

---

## Pointers vs Values

**Use pointers when:**
- ✅ Modifying the original data
- ✅ Large structs (avoid copying overhead)
- ✅ Need nil representation (optional values)
- ✅ Consistency (all methods or none on struct)

**Use values when:**
- ✅ Small structs (primitive-heavy, < 16 bytes)
- ✅ Immutable operations
- ✅ Map keys (must be comparable)
- ✅ Need to ensure data isn't modified

### Examples

```go
// GOOD - Pointer for modification
func (u *User) SetEmail(email string) {
    u.Email = email
}

// GOOD - Value for immutable read
func (u User) Validate() error {
    if u.Email == "" {
        return ErrInvalidEmail
    }
    return nil
}

// GOOD - Pointer for large struct
func ProcessOrder(order *Order) error {
    // Order is large, avoid copying
}

// GOOD - Value for small struct
type Point struct { X, Y int }
func Distance(p1, p2 Point) int {
    // Small struct, copying is cheap
}
```

---

## Generics vs interface{} vs any

**Use generics `[T any]` when:**
- ✅ Type-safe containers (slices, maps, trees)
- ✅ Type-safe utility functions (map, filter, reduce)
- ✅ Compile-time type checking required

**Use `interface{}` or `any` when:**
- ✅ Truly heterogeneous data (JSON parsing)
- ✅ Working with unknown external types
- ✅ Last resort when generics don't fit

### Examples

```go
// GOOD - Generic for type-safe container
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

// GOOD - Generic for utility
func MapSlice[T, U any](slice []T, mapper func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = mapper(v)
    }
    return result
}

// GOOD - interface{} for heterogeneous data
func HandleJSON(data []byte) error {
    var raw map[string]interface{}
    return json.Unmarshal(data, &raw)
}
```

---

## Return Value Conventions

**Multiple return values for errors:**
```go
func GetUser(id int64) (*User, error) {
    // Error as last return value
}
```

**Named return values (use sparingly):**
```go
func ConnectDB(dsn string) (db *gorm.DB, cleanup func(), err error) {
    // Named returns improve readability
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, nil, err
    }
    cleanup = func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()
    }
    return db, cleanup, nil
}
```

**Zero values for errors:**
```go
func FindUser(id int64) (User, error) {
    var user User
    err := db.Query(id, &user)
    return user, err
}
```

---

## var vs const

**Use `const` for:**
- ✅ Compile-time constants
- ✅ Configuration values that never change
- ✅ Magic numbers (give them meaning)

**Use `var` for:**
- ✅ Runtime values
- ✅ Pointers to immutable constants
- ✅ Variables that will be modified

### Examples

```go
// GOOD - const for immutable values
const (
    MaxRetries    = 3
    DefaultTimeout = 30 * time.Second
)

// GOOD - var for runtime values
var (
    db     *gorm.DB
    logger *zap.Logger
)

// GOOD - iota for enums
type Status int
const (
    StatusPending Status = iota
    StatusActive
    StatusCompleted
)
```

---

## Struct Design

**Embed vs composition:**
```go
// GOOD - Embed for "is-a" relationship
type Animal struct {
    Name string
}

type Dog struct {
    Animal  // Dog "is-a" Animal
    Breed string
}

// GOOD - Composition for "has-a" relationship
type Engine struct {
    Power int
}

type Car struct {
    engine *Engine  // Car "has-a" Engine
    model  string
}
```

**Constructor pattern:**
```go
func NewUser(name, email string) *User {
    return &User{
        Name:  name,
        Email: email,
    }
}

func NewDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return db, err
}
```

---

## Function Parameter Ordering

**Standard order (left to right):**
```
context → dependencies → config → input/output
```

### Examples

```go
// GOOD - Standard order
func (s *Service) CreateUser(
    ctx context.Context,           // 1. Context
    repo UserRepository,            // 2. Dependencies
    config CreateUserConfig,        // 3. Config
    req CreateUserRequest,         // 4. Input
) error

// GOOD - Constructor
func NewService(
    logger Logger,
    db Database,
    config *Config,
) *Service
```

**Variadic parameters come last:**
```go
func ProcessItems(ctx context.Context, items ...Item) error
```
