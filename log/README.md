# log

Zap-based structured logging with Sugar logger wrapper.

## Features

- Uber Zap Sugar logger
- Functional options pattern for configuration
- Console (colorized) and JSON encodings
- Global logger with package-level helpers
- Configurable log levels
- File output support
- Caller information and stacktraces

## Installation

```bash
go get github.com/haipham22/govern/log
```

## Usage

### Using Global Logger

```go
import "github.com/haipham22/govern/log"

// Simple logging
log.Info("Server starting")
log.Warn("High memory usage")
log.Error("Failed to connect")

// Formatted logging
log.Infof("Listening on %s", ":8080")

// Structured logging with key-value pairs
log.Infow("Server started", "port", 8080, "env", "production")
```

### Creating Custom Logger

```go
logger := log.New(
log.WithLevelString("debug"),
log.WithEncoding("json"),
log.WithOutputFile("app.log"),
)

logger.Info("Custom logger message")
```

### Setting Default Logger

```go
customLogger := log.New(log.WithLevelString("debug"))
log.SetDefault(customLogger)
```

## Options

| Option                | Description                        | Default   |
|-----------------------|------------------------------------|-----------|
| `WithLevel`           | Set log level (zapcore.Level)      | InfoLevel |
| `WithLevelString`     | Set log level from string          | InfoLevel |
| `WithEncoding`        | Set encoding ("json" or "console") | "console" |
| `WithTimeFormat`      | Set timestamp format               | ISO8601   |
| `WithOutput`          | Set output destination             | stdout    |
| `WithOutputFile`      | Set output to file                 | -         |
| `WithErrorOutput`     | Set error output                   | stderr    |
| `WithErrorOutputFile` | Set error output to file           | -         |
| `WithDevelopment`     | Enable development mode            | false     |

## Log Levels

| Level | String  |
|-------|---------|
| Debug | "debug" |
| Info  | "info"  |
| Warn  | "warn"  |
| Error | "error" |
| Fatal | "fatal" |
| Panic | "panic" |

## API

### Logger Creation

| Function                                 | Description            |
|------------------------------------------|------------------------|
| `New(opts ...Option) *zap.SugaredLogger` | Create new logger      |
| `Default() *zap.SugaredLogger`           | Get default logger     |
| `SetDefault(logger *zap.SugaredLogger)`  | Set default logger     |
| `Sync() error`                           | Flush buffered entries |

### Logging Functions

| Function                           | Description                        |
|------------------------------------|------------------------------------|
| `Debug(args ...interface{})`       | Log at debug level                 |
| `Info(args ...interface{})`        | Log at info level                  |
| `Warn(args ...interface{})`        | Log at warn level                  |
| `Error(args ...interface{})`       | Log at error level                 |
| `Fatal(args ...interface{})`       | Log and exit                       |
| `Debugf/Infof/Warnf/Errorf/Fatalf` | Formatted logging                  |
| `Debugw/Infow/Warnw/Errorw/Fatalw` | Structured logging with key-values |
