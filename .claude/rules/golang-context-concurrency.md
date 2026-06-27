# Go Context and Concurrency Rules

**Best practices for context, goroutines, channels, and concurrent operations in Go.**

---

## Context Usage

**ALWAYS pass context as first parameter:**
```go
// GOOD
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil {
        return err
    }
    return s.repo.Create(ctx, &user)
}

// BAD
func (s *Service) CreateUser(req CreateUserRequest) error {
    // Can't cancel or timeout this operation
    user, err := s.repo.FindByEmail(req.Email)
    return s.repo.Create(user)
}
```

**Context rules:**
- ✅ First parameter in functions that perform I/O
- ✅ Pass through call chain (handler → usecase → repository)
- ✅ Check `ctx.Err()` in long-running loops
- ✅ Use `context.WithTimeout()` for external calls
- ❌ NEVER store context in struct
- ❌ NEVER make context optional

### Example with timeout

```go
func (r *authRepository) FindByEmail(ctx context.Context, email string) (User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    var user User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    return user, err
}
```

---

## Goroutines Best Practices

**When to use goroutines:**
- ✅ Independent background tasks
- ✅ Parallel processing (map-reduce patterns)
- ✅ Fan-out/fan-in patterns
- ❌ NOT for simple sequential tasks

**ALWAYS manage goroutine lifecycle:**
```go
// GOOD - Controlled lifecycle
func (s *Service) ProcessBatch(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)
    
    for _, item := range items {
        item := item // Capture loop variable
        g.Go(func() error {
            return s.processItem(ctx, item)
        })
    }
    
    return g.Wait() // Wait for all or return first error
}

// BAD - Unbounded goroutines
func (s *Service) ProcessItems(items []Item) {
    for _, item := range items {
        go s.processItem(item) // May spawn thousands of goroutines
    }
}
```

**Worker pool pattern:**
```go
func workerPool(ctx context.Context, jobs <-chan Job, results chan<- Result, numWorkers int) {
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    results <- processJob(job)
                }
            }
        }()
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
}
```

---

## Channel Patterns

**Channel usage rules:**
- ✅ Use channels for goroutine communication
- ✅ Iterate over channels with `range`
- ✅ Close channels when done (sender side)
- ❌ NEVER send to closed channel
- ❌ NEVER close channel from receiver side

### Producer-consumer pattern

```go
// GOOD - Proper channel lifecycle
func Producer(ctx context.Context) <-chan Item {
    out := make(chan Item)
    
    go func() {
        defer close(out) // ALWAYS close sender channel
        
        for {
            select {
            case <-ctx.Done():
                return
            default:
                item := generateItem()
                out <- item
            }
        }
    }()
    
    return out
}

func Consumer(ctx context.Context, in <-chan Item) {
    for item := range in { // Iterate until channel closed
        select {
        case <-ctx.Done():
            return
        default:
            processItem(item)
        }
    }
}
```

**Buffered vs unbuffered:**
```go
// Unbuffered - Synchronous (default choice)
ch := make(chan Result)

// Buffered - Asynchronous (use with caution)
ch := make(chan Result, 100) // Use for: producer faster than consumer
```

---

## Common Concurrency Patterns

**Fan-out/fan-in:**
```go
func ProcessItems(items []Item) []Result {
    // Fan-out
    in := make(chan Item)
    go func() {
        for _, item := range items {
            in <- item
        }
        close(in)
    }()
    
    // Process in parallel
    out1 := make(chan Result)
    out2 := make(chan Result)
    go worker(in, out1)
    go worker(in, out2)
    
    // Fan-in
    var results []Result
    go func() {
        for r := range out1 {
            results = append(results, r)
        }
    }()
    go func() {
        for r := range out2 {
            results = append(results, r)
        }
    }()
    
    return results
}
```

**Pipeline:**
```go
func Pipeline(input <-chan int) <-chan int {
    step1 := make(chan int)
    step2 := make(chan int)
    
    // Stage 1
    go func() {
        for v := range input {
            step1 <- v * 2
        }
        close(step1)
    }()
    
    // Stage 2
    go func() {
        for v := range step1 {
            step2 <- v + 1
        }
        close(step2)
    }()
    
    return step2
}
```

---

## Safety Guidelines

**Race conditions:**
```go
// BAD - Race condition
var counter int

func increment() {
    counter++ // Not goroutine-safe
}

// GOOD - Use sync/atomic or mutex
var counter int64
var mu sync.Mutex

func increment() {
    atomic.AddInt64(&counter, 1)
}

// OR
func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

**Never do this:**
```go
// BAD - Starting goroutine without cleanup
func process() {
    go func() {
        // Abandoned goroutine
    }()
}

// BAD - Closing channel from receiver
func consumer(ch chan int) {
    for v := range ch {
        process(v)
    }
    close(ch) // WRONG: receiver shouldn't close
}
```
