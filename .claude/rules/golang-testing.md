# Go Testing Best Practices

**Best practices for writing effective, maintainable tests in Go.**

---

## Table-Driven Tests

**Use table-driven tests for multiple cases:**
```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"missing domain", "user@", true},
        {"empty", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Test Setup/Teardown

**Use setup functions for common logic:**
```go
func setupTest(t *testing.T) *Service {
    t.Helper()
    
    db := setupTestDB(t)
    repo := NewTestRepository(db)
    return NewService(repo)
}

func (s *Service) Close(t *testing.T) {
    t.Helper()
    s.db.Close()
}
```

---

## Subtests

**Use t.Run() for related tests:**
```go
func TestUserService(t *testing.T) {
    svc := setupTest(t)
    defer svc.Close(t)
    
    t.Run("CreateUser", func(t *testing.T) {
        // Test user creation
    })
    
    t.Run("UpdateUser", func(t *testing.T) {
        // Test user update
    })
    
    t.Run("DeleteUser", func(t *testing.T) {
        // Test user deletion
    })
}
```

---

## Test Helpers

**Use t.Helper() to skip helper frames:**
```go
func assertUser(t *testing.T, got, want User) {
    t.Helper() // Mark as helper
    
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

// Call in tests
func TestSomething(t *testing.T) {
    assertUser(t, user1, user2) // Failure reported here, not in assertUser
}
```

---

## Mocking Interfaces

**Use fakes or test doubles:**
```go
// Fake implementation for testing
type FakeUserRepository struct {
    users map[int64]User
}

func (f *FakeUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
    for _, user := range f.users {
        if user.Email == email {
            return user, nil
        }
    }
    return User{}, ErrUserNotFound
}

// Use in tests
func TestService(t *testing.T) {
    repo := &FakeUserRepository{
        users: map[int64]User{
            1: {ID: 1, Email: "test@example.com"},
        },
    }
    svc := NewService(repo)
    // Test...
}
```

---

## Race Detection

**ALWAYS run tests with race detector:**
```bash
go test -race ./...
```

**Test concurrent code:**
```go
func TestConcurrentAccess(t *testing.T) {
    svc := setupTest(t)
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int64) {
            defer wg.Done()
            svc.GetUser(id)
        }(int64(i))
    }
    
    wg.Wait()
    // No race conditions
}
```

---

## Test Coverage

**Check coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Set coverage threshold:**
```go
// Test package
func TestCoverage(t *testing.T) {
    if testing.CoverMode() == "" {
        t.Skip("skipping coverage test")
    }
    // Coverage tests...
}
```

---

## Benchmark Tests

**Write benchmarks for performance-critical code:**
```go
func BenchmarkUserCreation(b *testing.B) {
    svc := setupTest(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        svc.CreateUser(context.Background(), CreateUserRequest{
            Name:  "Test User",
            Email: "test@example.com",
        })
    }
}

// Run benchmarks
go test -bench=. -benchmem
```

---

## Example Tests

**Use examples for documentation:**
```go
// Example usage
func ExampleUser_Validate() {
    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    if err := user.Validate(); err != nil {
        fmt.Println(err)
        return
    }
    
    fmt.Println("User is valid")
    // Output: User is valid
}
```

---

## Test Data Management

**Use test data fixtures:**
```go
// testdata/users.go
var TestUsers = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

// Use in tests
func TestWithFixtures(t *testing.T) {
    for _, user := range TestUsers {
        // Test with fixture data
    }
}
```

---

## Integration Tests

**Use build tags for integration tests:**
```go
//go:build integration
// +build integration

func TestRealDatabase(t *testing.T) {
    // Integration test with real database
}

// Run integration tests
go test -tags=integration ./...
```

---

## Test Cleanup

**Cleanup after tests:**
```go
func TestWithCleanup(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()
    })
    
    // Test...
}
```

---

## Golden Files

**Use golden files for output verification:**
```go
func TestGoldenOutput(t *testing.T) {
    got := generateOutput()
    
    golden := filepath.Join("testdata", t.Name()+".golden")
    if *updateGolden {
        os.WriteFile(golden, got, 0644)
    }
    
    want, _ := os.ReadFile(golden)
    if !bytes.Equal(got, want) {
        t.Errorf("got %s, want %s", got, want)
    }
}
```
