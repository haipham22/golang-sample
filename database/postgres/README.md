# postgres

PostgreSQL connection management with GORM.

## Features

- GORM-based connections
- Functional options pattern for configuration
- Configurable connection pool settings
- Automatic cleanup functions
- Custom logger integration
- Simple protocol option for prepared statements

## Installation

```bash
go get github.com/haipham22/govern/database/postgres
```

## Usage

```go
package main

import (
    "log"
    "time"

    "github.com/haipham22/govern/database/postgres"
)

func main() {
    db, cleanup, err := postgres.New(
        "host=localhost user=postgres password=secret dbname=mydb",
        postgres.WithMaxOpenConns(50),
        postgres.WithConnMaxLifetime(10*time.Minute),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // Use db for database operations
    // ...
}
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithDebug` | Enable debug mode | `false` |
| `WithMaxIdleConns` | Maximum idle connections | `25` |
| `WithMaxOpenConns` | Maximum open connections | `100` |
| `WithConnMaxLifetime` | Connection max lifetime | `5min` |
| `WithConnMaxIdleTime` | Connection max idle time | `5min` |
| `WithPreferSimpleProtocol` | Disable prepared statements | `true` |
| `WithLogger` | Custom GORM logger | - |

## API

### `New(dsn string, opts ...Option) (*gorm.DB, func(), error)`

Creates a new GORM PostgreSQL connection.

**Returns:**
- `*gorm.DB`: GORM database instance
- `func()`: Cleanup function to close the connection
- `error`: Any error that occurred

## DSN Format

```go
// Connection string format
"host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable"

// Or URL format
"postgres://postgres:secret@localhost:5432/mydb?sslmode=disable"
```

## Connection Pool Defaults

| Setting | Value |
|---------|-------|
| MaxIdleConns | 25 |
| MaxOpenConns | 100 |
| ConnMaxLifetime | 5 minutes |
| ConnMaxIdleTime | 5 minutes |
