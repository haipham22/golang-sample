# database

Database client integrations for PostgreSQL and Redis.

## Overview

The `database` package provides database client integrations with connection pooling, configuration options, and cleanup functions for PostgreSQL and Redis.

## Subpackages

- `postgres` - PostgreSQL integration with GORM
- `redis` - Redis client integration

## postgres

### Key Functions

```go
// Create new PostgreSQL connection
func New(dsn string, opts ...Option) (*gorm.DB, func(), error)

// Configure connection pool
func ConfigureConnectionPool(db *gorm.DB, cfg *Config) (func(), error)
```

### Options

```go
func WithDebug(debug bool) Option
func WithMaxIdleConns(n int) Option
func WithMaxOpenConns(n int) Option
func WithConnMaxLifetime(d time.Duration) Option
func WithConnMaxIdleTime(d time.Duration) Option
func WithPreferSimpleProtocol(v bool) Option
func WithLogger(log logger.Interface) Option
```

### Defaults

```go
const (
    DefaultMaxIdleConns    = 25
    DefaultMaxOpenConns    = 100
    DefaultConnMaxLifetime = 5 * time.Minute
)
```

### Usage

```go
// Basic connection
dsn := "host=localhost user=postgres password=password dbname=mydb port=5432 sslmode=disable"
db, cleanup, err := postgres.New(dsn)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// With custom pool settings
db, cleanup, err := postgres.New(dsn,
    postgres.WithMaxOpenConns(50),
    postgres.WithMaxIdleConns(10),
    postgres.WithConnMaxLifetime(10*time.Minute),
)

// With debug mode
db, cleanup, err := postgres.New(dsn, postgres.WithDebug(true))

// Use GORM
var users []User
db.Find(&users)
```

## redis

### Key Functions

```go
// Create new Redis client
func New(addr string, opts ...Option) (redis.UniversalClient, func(), error)

// Create from DSN string
func NewFromDSN(dsn string, opts ...Option) (redis.UniversalClient, func(), error)

// Parse DSN string
func ParseDSN(dsn string) (string, []Option, error)
```

### Options

```go
func WithAddrs(addrs []string) Option
func WithPassword(password string) Option
func WithDB(db int) Option
func WithPoolSize(size int) Option
func WithMinIdleConns(n int) Option
func WithMaxRetries(n int) Option
func WithDialTimeout(d time.Duration) Option
func WithReadTimeout(d time.Duration) Option
func WithWriteTimeout(d time.Duration) Option
func WithPoolTimeout(d time.Duration) Option
```

### Defaults

```go
const (
    DefaultPoolSize = 100
    DefaultMinIdle  = 10
)
```

### Usage

```go
// Single node
client, cleanup, err := redis.New("localhost:6379")
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// With password and DB
client, cleanup, err := redis.New("localhost:6379",
    redis.WithPassword("password"),
    redis.WithDB(1),
)

// Cluster mode
client, cleanup, err := redis.New("",
    redis.WithAddrs([]string{":7000", ":7001", ":7002"}),
)

// From DSN
client, cleanup, err := redis.NewFromDSN("redis://:password@localhost:6379/1")

// Use client
ctx := context.Background()
err := client.Set(ctx, "key", "value", time.Hour).Err()
val, err := client.Get(ctx, "key").Result()
```

## DSN Format

Redis DSN format: `redis://[:password@]host[:port][/db][?options]`

Examples:
- `redis://localhost:6379` - Local Redis
- `redis://:password@localhost:6379/1` - With password and DB
- `redis://:password@host:6379/1?pool_size=100` - With options

## Connection Pool Management

Both packages return a cleanup function that should be called when done:

```go
db, cleanup, err := postgres.New(dsn)
if err != nil {
    return err
}
defer cleanup()  // Close connections when done
```

## Thread Safety

Both GORM and Redis clients are safe for concurrent use from multiple goroutines.

## References

- [database/postgres/postgres.go](../../database/postgres/postgres.go) - PostgreSQL implementation
- [database/redis/redis.go](../../database/redis/redis.go) - Redis implementation
- [database/redis/dsn.go](../../database/redis/dsn.go) - DSN parsing
