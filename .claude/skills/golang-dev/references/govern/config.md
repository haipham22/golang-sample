# govern/config

Import: `github.com/haipham22/govern/config`

Ultra-minimal config helper. YAML + env override + struct validation, generic API.

## Use When

- App loads config from YAML/env with validation.

## Basic

```go
import "github.com/haipham22/govern/config"

type Config struct {
    Server struct {
        Host string `validate:"required"`
        Port int    `validate:"required,min=1,max=65535"`
    } `validate:"required"`
}

cfg, err := config.Load[Config]("./config.yaml")
```

## With Options

```go
cfg, err := config.LoadWithOptions[Config](
    "./config.yaml",
    config.WithENVPrefix("APP"),
    config.WithLogger(logger),
)
```

## ENV Override

`SECTION_KEY` → `section.key`. With prefix `APP`: `APP_SERVER_PORT` → `server.port`.

```bash
APP_SERVER_PORT=9090 ./app
```

## Validation

go-playground/validator tags:

| Tag | Meaning |
|---|---|
| `required` | Field must be present |
| `min=X` / `max=X` | Min/max (length or value) |
| `oneof=a b c` | Enum |
| `hostname` / `ip` / `email` | Format |

## Rules

- ✅ Define one typed config struct; load once at composition root.
- ✅ Pass typed values into constructors, not whole config bag.
- ✅ Fail fast on missing/invalid config.
- ✅ Store secrets (DSN, JWT secret) in env, never in committed YAML.
- ❌ Do not call viper/env parsing inside services or repositories.
- ❌ Do not split DSN into host/port/user/password fields — keep one string.

## Avoid

- Raw Viper boilerplate + separate validation pass.
- Per-package env parsing.

## Reference

Source: [`config/`](../../../../../../../config/).
