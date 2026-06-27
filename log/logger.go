package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap.SugaredLogger with the given options.
// If no options are provided, it uses sensible defaults:
// - Level: Info
// - Encoding: console (human-readable)
// - Output: stdout
func New(opts ...Option) *zap.SugaredLogger {
	cfg := &config{
		level:       zapcore.InfoLevel,
		encoding:    "console",
		timeFormat:  DefaultTimeFormat,
		output:      zapcore.AddSync(os.Stdout),
		errorOutput: zapcore.AddSync(os.Stderr),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.encoding == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if cfg.encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		cfg.output,
		cfg.level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger.Sugar()
}

// Default returns the default global logger.
func Default() *zap.SugaredLogger {
	return defaultLogger
}

// SetDefault sets the default global logger.
func SetDefault(logger *zap.SugaredLogger) {
	defaultLogger = logger
}

// Sync flushes any buffered log entries.
func Sync() error {
	return defaultLogger.Sync()
}
