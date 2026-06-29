# govern/log

Import: `github.com/haipham22/govern/log`

Zap Sugar logger wrapper with functional options. Console (colorized) + JSON encodings, global logger helpers, file output.

## Use When

- App needs structured logging.

## Global Logger

```go
import "github.com/haipham22/govern/log"

log.Info("server starting")
log.Infof("listening on %s", ":8080")
log.Infow("server started", "port", 8080, "env", "production")
```

## Custom Logger (inject)

```go
logger := log.New(
    log.WithLevelString("debug"),
    log.WithEncoding("json"),
    log.WithOutputFile("app.log"),
)
```

## Options

| Option | Description | Default |
|---|---|---|
| `WithLevel(lvl)` | zapcore.Level | InfoLevel |
| `WithLevelString(s)` | Level from string | InfoLevel |
| `WithEncoding(s)` | "json" or "console" | "console" |
| `WithTimeFormat(s)` | Timestamp format | ISO8601 |
| `WithOutput(w)` | Output destination | stdout |
| `WithOutputFile(path)` | File output | - |
| `WithErrorOutput(w)` | Error output | stderr |
| `WithErrorOutputFile(path)` | Error file output | - |
| `WithDevelopment(b)` | Dev mode | false |

## Levels

`"debug"`, `"info"`, `"warn"`, `"error"`, `"fatal"`, `"panic"`.

## API

| Function | Description |
|---|---|
| `New(opts...) *zap.SugaredLogger` | Create logger |
| `Default() *zap.SugaredLogger` | Get default |
| `SetDefault(logger)` | Set default |
| `Sync() error` | Flush buffered entries |
| `Debug/Info/Warn/Error/Fatal(args...)` | Level logging |
| `Debugf/Infof/Warnf/Errorf/Fatalf(fmt, args...)` | Formatted |
| `Debugw/Infow/Warnw/Errorw/Fatalw(msg, kvs...)` | Structured k-v |

## Rules

- ✅ Inject `*zap.SugaredLogger` into constructors; avoid globals in business logic.
- ✅ Use structured `*w` calls with key-value pairs.
- ✅ Add request context fields (request_id, user_id).
- ✅ Use correct level: Debug (dev), Info (lifecycle), Warn (recoverable), Error (needs attention), Fatal (stop).
- ❌ Never log secrets, raw DSNs, passwords, tokens, PII.
- ❌ Never use stdlib `log.Printf` for service logging.
- ❌ Avoid logging in hot loops (perf).

## Avoid

- Configuring Zap by hand in every binary.
- Re-deriving Sugar logger when govern already provides it.

## Reference

Source: [`log/`](../../../../../../../log/).
