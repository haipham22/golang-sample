# govern/database/postgres

Import: `github.com/haipham22/govern/database/postgres`

PostgreSQL connection via GORM. Functional options, pool defaults, cleanup triple.

## Use When

- App connects to Postgres.

## New

```go
import "github.com/haipham22/govern/database/postgres"

db, cleanup, err := postgres.New(
    "host=localhost user=postgres password=secret dbname=mydb sslmode=disable",
    postgres.WithMaxOpenConns(50),
    postgres.WithConnMaxLifetime(10*time.Minute),
)
if err != nil { log.Fatal(err) }
defer cleanup()
```

Signature: `New(dsn string, opts ...Option) (*gorm.DB, func(), error)`.

## Options

| Option | Description | Default |
|---|---|---|
| `WithDebug` | Enable GORM debug | false |
| `WithMaxIdleConns` | Max idle conns | 25 |
| `WithMaxOpenConns` | Max open conns | 100 |
| `WithConnMaxLifetime` | Conn max lifetime | 5m |
| `WithConnMaxIdleTime` | Conn max idle time | 5m |
| `WithPreferSimpleProtocol` | Disable prepared statements | true |
| `WithLogger` | Custom GORM logger | - |

## DSN Format

```go
"host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable"
"postgres://postgres:secret@localhost:5432/mydb?sslmode=disable"
```

## Rules

- ✅ Capture `(db, cleanup, err)`; `defer cleanup()` immediately.
- ✅ Build once in bootstrap; pass `*gorm.DB` down.
- ✅ Use `db.WithContext(ctx)` on every query.
- ✅ Keep DSN as one string — let driver parse.
- ✅ Use `sslmode=require` in production.
- ❌ Never log raw DSN (contains password).
- ❌ Never `gorm.Open` directly at call sites.
- ❌ Do not decompose DSN into host/port/user/password config fields.

## Avoid

- Direct `gorm.Open` (no cleanup triple, no pool defaults).
- Building DSN via `fmt.Sprintf` at every call site.

## Reference

Source: [`database/postgres/`](../../../../../../../database/postgres/). Uses `gorm.io/driver/postgres`, `github.com/jackc/pgx/v5`. See also [config DSN rule](../../development-rules.md).
