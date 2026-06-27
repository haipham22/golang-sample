# redis

Redis connection management with go-redis/v9. Supports both single-node and cluster configurations.

## Features

- go-redis/v9 UniversalClient for standalone and cluster modes
- Functional options pattern for configuration
- DSN string parsing support
- Configurable connection pool settings
- Automatic cleanup functions
- Cluster support with latency-based and random routing

## Installation

```bash
go get github.com/haipham22/govern/database/redis
```

## Usage

### Simple Connection

```go
package main

import (
    "log"

    "github.com/haipham22/govern/database/redis"
)

func main() {
    client, cleanup, err := redis.New("localhost:6379")
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // Use client
    _ = client
}
```

### Cluster Connection

```go
client, cleanup, err := redis.New("",
    redis.WithAddrs("node1:6379", "node2:6379", "node3:6379"),
    redis.WithRouteByLatency(true),
)
if err != nil {
    log.Fatal(err)
}
defer cleanup()
```

### DSN Connection

```go
// redis://[:password@]host[:port][/db][?options]
client, cleanup, err := redis.NewFromDSN("redis://:secret@localhost:6379/1?pool_size=50")
```

## Options

| Option                | Description              | Default |
|-----------------------|--------------------------|---------|
| `WithAddr`            | Server address           | -       |
| `WithPassword`        | Authentication password  | -       |
| `WithDB`              | Database number          | `0`     |
| `WithPoolSize`        | Connection pool size     | `100`   |
| `WithMinIdleConns`    | Minimum idle connections | `10`    |
| `WithMaxRetries`      | Maximum retry attempts   | `3`     |
| `WithMaxRetryBackoff` | Maximum retry backoff    | `500ms` |
| `WithMinRetryBackoff` | Minimum retry backoff    | `8ms`   |
| `WithDialTimeout`     | Dial timeout             | `5s`    |
| `WithReadTimeout`     | Read timeout             | `3s`    |
| `WithWriteTimeout`    | Write timeout            | `3s`    |
| `WithPoolTimeout`     | Pool timeout             | `4s`    |
| `WithConnMaxIdleTime` | Idle connection timeout  | `5m`    |
| `WithAddrs`           | Cluster addresses         | -       |
| `WithRouteByLatency`  | Route by latency         | `false` |
| `WithRouteRandomly`   | Random shard selection   | `false` |

## DSN Format

```
redis://[:password@]host[:port][/db][?options]
rediss://[:password@]host[:port][/db][?options]  # TLS
```

### Query Options

- `db`: database number
- `pool_size`: connection pool size
- `min_idle`: minimum idle connections
- `max_retries`: maximum retry attempts
- `dial_timeout`: dial timeout in seconds
- `read_timeout`: read timeout in seconds
- `write_timeout`: write timeout in seconds
- `pool_timeout`: pool timeout in seconds
- `idle_timeout`: idle timeout in seconds

### DSN Examples

```go
"redis://localhost:6379"
"redis://:secret@localhost:6379"
"redis://localhost:6379/1"
"redis://:secret@localhost:6379/2?pool_size=50&min_idle=5"
"redis://localhost:6379?dial_timeout=10&read_timeout=5"
```

## API

### `New(addr string, opts ...Option) (redis.UniversalClient, func(), error)`

Creates a new Redis client with explicit address. Returns UniversalClient which supports both standalone and cluster modes.

### `NewFromDSN(dsn string, opts ...Option) (redis.UniversalClient, func(), error)`

Creates a new Redis client from a DSN string. Returns UniversalClient interface.

## Connection Pool Defaults

| Setting | Value |
|---------|-------|
| PoolSize | 100 |
| MinIdleConns | 10 |
| MaxRetries | 3 |
| DialTimeout | 5 seconds |
| ReadTimeout | 3 seconds |
| WriteTimeout | 3 seconds |
| PoolTimeout | 4 seconds |
| ConnMaxIdleTime | 5 minutes |
