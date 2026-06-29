# govern/database/redis

Import: `github.com/haipham22/govern/database/redis`

Redis via go-redis/v9 `UniversalClient` (standalone + cluster). Functional options, DSN parsing, cleanup triple.

## Use When

- App connects to Redis (standalone or cluster).

## New (address)

```go
import "github.com/haipham22/govern/database/redis"

client, cleanup, err := redis.New("localhost:6379")
if err != nil { log.Fatal(err) }
defer cleanup()
```

Signature: `New(addr string, opts ...Option) (redis.UniversalClient, func(), error)`.

## NewFromDSN

```go
client, cleanup, err := redis.NewFromDSN("redis://:secret@localhost:6379/1?pool_size=50")
```

DSN: `redis://[:password@]host[:port][/db][?options]`, `rediss://...` for TLS.

Query options: `db`, `pool_size`, `min_idle`, `max_retries`, `dial_timeout`, `read_timeout`, `write_timeout`, `pool_timeout`, `idle_timeout`.

## Cluster

```go
client, cleanup, err := redis.New("",
    redis.WithAddrs("node1:6379", "node2:6379", "node3:6379"),
    redis.WithRouteByLatency(true),
)
```

## Options

| Option | Description | Default |
|---|---|---|
| `WithAddr` | Server address | - |
| `WithPassword` | Auth password | - |
| `WithDB` | Database number | 0 |
| `WithPoolSize` | Pool size | 100 |
| `WithMinIdleConns` | Min idle | 10 |
| `WithMaxRetries` | Max retries | 3 |
| `WithDialTimeout` | Dial timeout | 5s |
| `WithReadTimeout` / `WithWriteTimeout` | I/O timeouts | 3s |
| `WithAddrs(...)` | Cluster addresses | - |
| `WithRouteByLatency` / `WithRouteRandomly` | Cluster routing | false |

## Rules

- ✅ Capture `(client, cleanup, err)`; `defer cleanup()`.
- ✅ Build once in bootstrap; pass `UniversalClient` down.
- ✅ Pass context to Redis ops.
- ✅ Use `rediss://` (TLS) in production.
- ❌ Never log raw URL (may contain password).
- ❌ Never `redis.NewClient` directly at call sites.

## Avoid

- Manual URL parsing.
- Per-handler Redis clients.

## Reference

Source: [`database/redis/`](../../../../../../../database/redis/). Uses `github.com/redis/go-redis/v9`.
