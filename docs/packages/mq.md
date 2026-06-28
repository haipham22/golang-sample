# mq

Message queue integration with Asynq for background task processing.

## Overview

The `mq` package provides Asynq integration for background task processing with Govern patterns for graceful shutdown and lifecycle management.

## Subpackages

- `asynq` - Asynq task queue server and handlers

## asynq

### Key Types

### Server

```go
type Server struct {
    server    *asynq.Server
    mux       *TaskMux
    logger    *zap.SugaredLogger
    config    *Config
}
```

### TaskMux

```go
type TaskMux struct {
    handlers map[string]HandlerFunc
    mu       sync.RWMutex
}
```

### HandlerFunc

```go
type HandlerFunc func(context.Context, *asynq.Task) error
```

## Key Functions

### Server Functions

```go
// Create new server
func NewServer(redisClient redis.UniversalClient, mux *TaskMux, opts ...Option) (*Server, func(), error)

// Start processing tasks (implements graceful.Service)
func (s *Server) Start(ctx context.Context) error

// Shutdown gracefully (implements graceful.Service)
func (s *Server) Shutdown(ctx context.Context) error

// Close server connection
func (s *Server) Close() error
```

### TaskMux Functions

```go
// Create new task multiplexer
func NewTaskMux() *TaskMux

// Register handler for task type
func (m *TaskMux) HandleFunc(pattern string, handler HandlerFunc)

// Implement asynq.Handler interface
func (m *TaskMux) HandleTask(ctx context.Context, t *asynq.Task) error
```

### Options

```go
func WithConcurrency(n int) Option
func WithQueues(queues map[string]int) Option
func WithLogger(logger *zap.SugaredLogger) Option
func WithShutdownTimeout(timeout time.Duration) Option
```

## Usage

### Basic Server Setup

```go
import "github.com/haipham22/govern/mq/asynq"
import "github.com/redis/go-redis/v9"

// Create Redis client
redisClient, cleanup, err := redis.New("localhost:6379")
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// Create task multiplexer
mux := asynq.NewTaskMux()

// Register task handlers
mux.HandleFunc("email:send", handleSendEmail)
mux.HandleFunc("image:process", handleProcessImage)

// Create server
server, cleanup, err := asynq.NewServer(redisClient, mux,
    asynq.WithConcurrency(10),
    asynq.WithLogger(logger),
)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// Start server (blocks until shutdown)
err = server.Start(context.Background())
```

### Task Handler

```go
func handleSendEmail(ctx context.Context, t *asynq.Task) error {
    // Parse task payload
    var payload EmailPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }

    // Send email
    if err := sendEmail(payload.To, payload.Subject, payload.Body); err != nil {
        return fmt.Errorf("send email failed: %w", err)
    }

    log.Infof("Email sent to %s", payload.To)
    return nil
}

type EmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

### Enqueuing Tasks

```go
import "github.com/hibiken/asynq"

// Create client
client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

// Enqueue task
task := asynq.NewTask("email:send", map[string]interface{}{
    "to":       "user@example.com",
    "subject":  "Welcome",
    "body":     "Welcome to our service!",
})

info, err := client.Enqueue(task)
if err != nil {
    log.Fatal(err)
}

log.Infof("Task enqueued: %s", info.ID)
```

### With Graceful Shutdown

```go
import "github.com/haipham22/govern/graceful"

server, cleanup, err := asynq.NewServer(redisClient, mux)
defer cleanup()

// Use with graceful.Run
graceful.Run(ctx, logger, 30*time.Second, server)
```

### Custom Queues

```go
// Configure queues with priorities
queues := map[string]int{
    "critical": 6,
    "default":  3,
    "low":      1,
}

server, cleanup, err := asynq.NewServer(redisClient, mux,
    asynq.WithQueues(queues),
)

// Enqueue to specific queue
task := asynq.NewTask("email:send", payload, asynq.Queue("critical"))
client.Enqueue(task)
```

### Task Retry

```go
// Task with retry options
task := asynq.NewTask("email:send", payload,
    asynq.MaxRetry(5),
    asynq.Timeout(10*time.Minute),
)

client.Enqueue(task)
```

### Scheduled Tasks

```go
// Process task in 1 hour
task := asynq.NewTask("email:send", payload)
info, err := client.Enqueue(task, asynq.ProcessIn(time.Hour))

// Process at specific time
task = asynq.NewTask("email:send", payload)
info, err = client.Enqueue(task, asynq.ProcessAt(time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC))

// Run every hour
task = asynq.NewTask("cleanup:logs", payload)
info, err = client.Enqueue(task, asynq.ProcessAt(time.Now()), asynq.WrapAsynqCronJob("0 * * * *"))
```

### Task Groups

```go
// Create task group
group := asynq.NewGroup("email:send", []*asynq.Task{
    asynq.NewTask("email:send", email1),
    asynq.NewTask("email:send", email2),
    asynq.NewTask("email:send", email3),
})

// Enqueue group
info, err := client.EnqueueGroup(group, asynq.GroupSize(2))
```

## Task Handler Patterns

### Simple Handler

```go
func handleSimple(ctx context.Context, t *asynq.Task) error {
    log.Infof("Processing task: %s", t.Type())
    // Process task
    return nil
}
```

### With Error Recovery

```go
func handleWithRecovery(ctx context.Context, t *asynq.Task) error {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("Task panic: %v", r)
        }
    }()

    return processTask(ctx, t)
}
```

### With Timeout

```go
func handleWithTimeout(ctx context.Context, t *asynq.Task) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    return processTask(ctx, t)
}
```

### With Deadlines

```go
func handleWithDeadline(ctx context.Context, t *asynq.Task) error {
    deadline, ok := t.Deadline()
    if ok && time.Until(deadline) < 30*time.Second {
        return fmt.Errorf("insufficient time remaining: %v", deadline)
    }

    return processTask(ctx, t)
}
```

## Server Configuration

### Default Configuration

```go
Default Concurrency: 10
Default Queues:      {"default": 1}
Default Timeout:     N/A
```

### Custom Configuration

```go
server, cleanup, err := asynq.NewServer(redisClient, mux,
    asynq.WithConcurrency(20),
    asynq.WithQueues(map[string]int{
        "critical": 10,
        "default":  5,
        "low":      2,
    }),
    asynq.WithShutdownTimeout(60*time.Second),
)
```

## Best Practices

1. **Idempotent handlers** - Tasks should be safe to retry
2. **Timeout handling** - Respect context cancellation
3. **Error handling** - Return errors for retries
4. **Logging** - Log task processing for debugging
5. **Monitoring** - Track task processing metrics

## Monitoring

### With Metrics

```go
import "github.com/haipham22/govern/metrics"

var (
    tasksProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "asynq_tasks_processed_total",
            Help: "Total number of tasks processed",
        },
        []string{"type", "status"},
    )
)

func handleWithMetrics(ctx context.Context, t *asynq.Task) error {
    start := time.Now()
    err := processTask(ctx, t)
    
    status := "success"
    if err != nil {
        status = "error"
    }
    
    tasksProcessed.WithLabelValues(t.Type(), status).Inc()
    return err
}
```

## References

- [mq/asynq/server.go](../../mq/asynq/server.go) - Server implementation
- [mq/asynq/handler.go](../../mq/asynq/handler.go) - Handler implementation
- [mq/asynq/config.go](../../mq/asynq/config.go) - Configuration options
