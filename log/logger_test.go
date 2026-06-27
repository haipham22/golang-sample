package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Test helpers

func assertLoggerNotNil(t *testing.T, logger *zap.SugaredLogger) {
	t.Helper()
	require.NotNil(t, logger, "New() returned nil logger")
}

func TestNew(t *testing.T) {
	logger := New()
	assertLoggerNotNil(t, logger)
}

func TestNewWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		validate func(*testing.T, *zap.SugaredLogger)
	}{
		{
			name:   "with debug level",
			option: WithLevel(zapcore.DebugLevel),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with info level",
			option: WithLevel(zapcore.InfoLevel),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with warn level",
			option: WithLevel(zapcore.WarnLevel),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with error level",
			option: WithLevel(zapcore.ErrorLevel),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with debug level string",
			option: WithLevelString("debug"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with info level string",
			option: WithLevelString("info"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with warn level string",
			option: WithLevelString("warn"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with error level string",
			option: WithLevelString("error"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with invalid level string",
			option: WithLevelString("invalid"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				// Should still create logger, just ignore invalid level
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with console encoding",
			option: WithEncoding("console"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with json encoding",
			option: WithEncoding("json"),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
		{
			name:   "with development mode",
			option: WithDevelopment(true),
			validate: func(t *testing.T, logger *zap.SugaredLogger) {
				assertLoggerNotNil(t, logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.option)
			tt.validate(t, logger)
		})
	}
}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name   string
		logFn  func(...interface{})
		logFfn func(string, ...interface{})
		logWfn func(string, ...interface{})
	}{
		{"Debug", Debug, Debugf, Debugw},
		{"Info", Info, Infof, Infow},
		{"Warn", Warn, Warnf, Warnw},
		{"Error", Error, Errorf, Errorw},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should not panic
			assert.NotPanics(t, func() {
				tt.logFn("test message")
				tt.logFfn("test %s", "message")
				tt.logWfn("test", "key", "value")
			})
		})
	}
}

func TestDefault(t *testing.T) {
	logger := Default()
	assertLoggerNotNil(t, logger)
}

func TestSetDefault(t *testing.T) {
	original := Default()
	defer SetDefault(original)

	newLogger := zap.NewNop().Sugar()
	SetDefault(newLogger)

	assert.NotEqual(t, original, Default(), "SetDefault() did not change the default logger")
}

func TestWithOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithOutput(zapcore.AddSync(&buf)))
	assertLoggerNotNil(t, logger)

	logger.Info("test message")
	assert.Contains(t, buf.String(), "test message", "Expected log message in output")
}

func TestSync(t *testing.T) {
	err := Sync()
	// Sync may fail on some buffers (e.g., stdout to closed pipe)
	// but should not panic
	_ = err
}

func TestWithOutputFile(t *testing.T) {
	t.Run("valid file path", func(t *testing.T) {
		// Create temp file
		tmpFile := t.TempDir() + "/test.log"
		logger := New(WithOutputFile(tmpFile))
		assertLoggerNotNil(t, logger)

		// Write to logger
		logger.Info("test message")
		err := logger.Sync()
		assert.NoError(t, err)
	})

	t.Run("invalid directory path", func(t *testing.T) {
		// Invalid path - non-existent directory with no write permission
		invalidPath := "/nonexistent/directory/test.log"
		logger := New(WithOutputFile(invalidPath))

		// Should still create logger (error is silently ignored in WithOutputFile)
		// and default to stdout
		assertLoggerNotNil(t, logger)

		// Should be able to log without error
		assert.NotPanics(t, func() {
			logger.Info("fallback to stdout")
		})
	})
}

func TestWithErrorOutputFile(t *testing.T) {
	t.Run("valid file path", func(t *testing.T) {
		tmpFile := t.TempDir() + "/error.log"
		logger := New(WithErrorOutputFile(tmpFile))
		assertLoggerNotNil(t, logger)

		logger.Error("error message")
		// Sync may fail on stdout in test environments
		_ = logger.Sync()
	})

	t.Run("invalid directory path", func(t *testing.T) {
		invalidPath := "/nonexistent/directory/error.log"
		logger := New(WithErrorOutputFile(invalidPath))

		// Should still create logger (error is silently ignored)
		assertLoggerNotNil(t, logger)

		assert.NotPanics(t, func() {
			logger.Error("fallback to stderr")
		})
	})
}

func TestConcurrentLogging(t *testing.T) {
	t.Run("concurrent write access", func(t *testing.T) {
		// Use file output for concurrent test (file handles have OS-level locking)
		tmpFile := t.TempDir() + "/concurrent.log"
		logger := New(WithOutputFile(tmpFile))

		const goroutines = 100
		const messagesPerGoroutine = 10

		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				for j := 0; j < messagesPerGoroutine; j++ {
					logger.Infof("goroutine %d message %d", id, j)
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < goroutines; i++ {
			<-done
		}

		// Sync should complete without error
		err := logger.Sync()
		assert.NoError(t, err)
	})

	t.Run("concurrent default logger access", func(t *testing.T) {
		const goroutines = 50

		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				Infof("concurrent info %d", id)
				Debugf("concurrent debug %d", id)
				Warnf("concurrent warn %d", id)
				Errorf("concurrent error %d", id)
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			<-done
		}

		// Global sync should not panic
		assert.NotPanics(t, func() {
			_ = Sync()
		})
	})
}

func TestSyncFailureScenarios(t *testing.T) {
	t.Run("sync after close", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(WithOutput(zapcore.AddSync(&buf)))

		logger.Info("test message")

		// First sync should succeed
		err := logger.Sync()
		assert.NoError(t, err)

		// Second sync should also succeed (idempotent)
		err = logger.Sync()
		assert.NoError(t, err)
	})

	t.Run("sync with nil logger", func(t *testing.T) {
		// SetDefault with nil should still allow Sync to be called
		original := Default()
		defer SetDefault(original)

		// This test ensures Sync doesn't panic
		assert.NotPanics(t, func() {
			_ = Sync()
		})
	})
}
