# Common Go Idioms

**Idiomatic Go patterns and best practices for everyday coding.**

---

## Interface Satisfaction

**No need to explicitly implement interfaces:**
```go
type UserRepository interface {
    FindByEmail(ctx context.Context, email string) (User, error)
    Create(ctx context.Context, user *User) error
}

// ANY type with these methods satisfies the interface
type postgresUserRepository struct {
    db *gorm.DB
}

// Automatically satisfies UserRepository interface
func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    return user, err
}
```

---

## Defer for Cleanup

**ALWAYS use defer for resource cleanup:**
```go
// GOOD - Defer cleanup
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // ALWAYS close resources
    
    // Process file...
    return nil
}

// GOOD - Defer with error handling
func processWithTx(db *gorm.DB) error {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if err := process1(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

---

## Zero Values

**Use zero values for defaults:**
```go
// Use zero values for defaults
var (
    defaultUser = User{}       // Zero value
    defaultConfig = Config{}   // Zero value
)

// Check for zero value
if user == (User{}) {
    // User not initialized
}

// Pointer vs value
func UpdateUser(user *User) error {
    if user == nil {
        return ErrInvalidUser
    }
    // user can be checked for nil
}
```

---

## Method Receivers

**Use value receivers for immutable:**
```go
// Use value receivers for:
// - Immutable operations
// - Small structs
// - When value type needs to satisfy interface

func (u User) Validate() error {
    // Cannot modify u
    return nil
}

// Use pointer receivers for:
// - Mutable operations
// - Large structs
// - Consistency

func (u *User) Save() error {
    u.UpdatedAt = time.Now() // Can modify u
    return nil
}
```

---

## Range Over Channels

**ALWAYS iterate over channels with range:**
```go
// GOOD - Range over channel
for result := range resultsChannel {
    process(result)
}

// BAD - Manual channel iteration
for {
    result, ok := <-resultsChannel
    if !ok {
        break
    }
    process(result)
}
```

---

## String Building

**Use strings.Builder for concatenation:**
```go
// GOOD - strings.Builder
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item.Name)
    builder.WriteString(",")
}
result := builder.String()

// BAD - String concatenation in loop
result := ""
for _, item := range items {
    result += item.Name + "," // Inefficient
}
```

---

## Error Wrapping Chain

**Don't over-wrap errors:**
```go
// GOOD - Minimal wrapping
if err := db.Create(user).Error; err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// BAD - Excessive wrapping
if err := db.Create(user).Error; err != nil {
    return fmt.Errorf("create user failed: user %v: %w", user, err)
}
```

---

## Slice Growth

**Preallocate when size is known:**
```go
// GOOD - Preallocate
users := make([]User, 0, 100) // Capacity 100
for _, u := range inputUsers {
    users = append(users, u)
}

// BAD - No preallocation
var users []User
for _, u := range inputUsers {
    users = append(users, u) // Multiple reallocations
}
```

---

## Type Assertions

**Use comma-ok pattern:**
```go
// GOOD - Comma-ok assertion
if v, ok := someValue.(User); ok {
    // Use v as User
}

// BAD - Panic on failure
v := someValue.(User) // Panics if not User
```

---

## Select Over Time

**Use select for timeouts:**
```go
// GOOD - Select with timeout
select {
case result := <-results:
    process(result)
case <-time.After(5 * time.Second):
    return ErrTimeout
}

// GOOD - Select with context
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-results:
    process(result)
}
```
