package asynq

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWithConcurrency(t *testing.T) {
	cfg := defaultConfig()
	opt := WithConcurrency(20)
	opt(cfg)

	if cfg.Concurrency != 20 {
		t.Errorf("expected Concurrency 20, got %d", cfg.Concurrency)
	}
}

func TestWithConcurrencyZero(t *testing.T) {
	cfg := defaultConfig()
	original := cfg.Concurrency
	opt := WithConcurrency(0)
	opt(cfg)

	if cfg.Concurrency != original {
		t.Errorf("expected Concurrency %d (unchanged), got %d", original, cfg.Concurrency)
	}
}

func TestWithQueues(t *testing.T) {
	cfg := defaultConfig()
	queues := map[string]int{"critical": 5, "default": 1, "low": 1}
	opt := WithQueues(queues)
	opt(cfg)

	if len(cfg.Queues) != 3 {
		t.Errorf("expected 3 queues, got %d", len(cfg.Queues))
	}

	if cfg.Queues["critical"] != 5 {
		t.Errorf("expected critical queue priority 5, got %d", cfg.Queues["critical"])
	}
}

func TestWithQueuesEmpty(t *testing.T) {
	cfg := defaultConfig()
	originalLen := len(cfg.Queues)
	opt := WithQueues(map[string]int{})
	opt(cfg)

	if len(cfg.Queues) != originalLen {
		t.Errorf("expected queues unchanged (len %d), got %d", originalLen, len(cfg.Queues))
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	cfg := defaultConfig()
	opt := WithShutdownTimeout(60 * time.Second)
	opt(cfg)

	expected := 60 * time.Second
	if cfg.ShutdownTimeout != expected {
		t.Errorf("expected ShutdownTimeout %v, got %v", expected, cfg.ShutdownTimeout)
	}
}

func TestWithClientMaxRetry(t *testing.T) {
	cfg := defaultClientConfig()
	opt := WithClientMaxRetry(10)
	opt(cfg)

	if cfg.MaxRetry != 10 {
		t.Errorf("expected MaxRetry 10, got %d", cfg.MaxRetry)
	}
}

func TestWithClientTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	opt := WithClientTimeout(5 * time.Minute)
	opt(cfg)

	expected := 5 * time.Minute
	if cfg.Timeout != expected {
		t.Errorf("expected Timeout %v, got %v", expected, cfg.Timeout)
	}
}

func TestWithLogger(t *testing.T) {
	cfg := defaultConfig()
	logger := zap.NewNop().Sugar()
	opt := WithLogger(logger)
	opt(cfg)

	// Logger option is a no-op in config (stored at package level)
	// This test ensures it doesn't panic
	if cfg == nil {
		t.Error("config should not be nil")
	}
}

func TestWithQueue(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithQueue("critical")
	opt(cfg)

	if cfg.Queue != "critical" {
		t.Errorf("expected Queue critical, got %s", cfg.Queue)
	}
}

func TestWithQueueEmpty(t *testing.T) {
	cfg := defaultEnqueueConfig()
	original := cfg.Queue
	opt := WithQueue("")
	opt(cfg)

	if cfg.Queue != original {
		t.Errorf("expected Queue unchanged (%s), got %s", original, cfg.Queue)
	}
}

func TestWithMaxRetry(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithMaxRetry(5)
	opt(cfg)

	if cfg.MaxRetry != 5 {
		t.Errorf("expected MaxRetry 5, got %d", cfg.MaxRetry)
	}
}

func TestWithMaxRetryZero(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithMaxRetry(0)
	opt(cfg)

	if cfg.MaxRetry != 0 {
		t.Errorf("expected MaxRetry 0, got %d", cfg.MaxRetry)
	}
}

func TestWithEnqueueTimeout(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithEnqueueTimeout(10 * time.Minute)
	opt(cfg)

	expected := 10 * time.Minute
	if cfg.Timeout != expected {
		t.Errorf("expected Timeout %v, got %v", expected, cfg.Timeout)
	}
}

func TestWithEnqueueTimeoutZero(t *testing.T) {
	cfg := defaultEnqueueConfig()
	original := cfg.Timeout
	opt := WithEnqueueTimeout(0)
	opt(cfg)

	if cfg.Timeout != original {
		t.Errorf("expected Timeout unchanged (%v), got %v", original, cfg.Timeout)
	}
}

func TestWithDeadline(t *testing.T) {
	cfg := defaultEnqueueConfig()
	now := time.Now().Add(1 * time.Hour)
	opt := WithDeadline(now)
	opt(cfg)

	if !cfg.Deadline.Equal(now) {
		t.Errorf("expected Deadline %v, got %v", now, cfg.Deadline)
	}
}

func TestWithDeadlineZero(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithDeadline(time.Time{})
	opt(cfg)

	if !cfg.Deadline.IsZero() {
		t.Errorf("expected Deadline zero, got %v", cfg.Deadline)
	}
}

func TestWithUnique(t *testing.T) {
	cfg := defaultEnqueueConfig()
	ttl := 5 * time.Minute
	opt := WithUnique(ttl)
	opt(cfg)

	if cfg.UniqueTTL != ttl {
		t.Errorf("expected UniqueTTL %v, got %v", ttl, cfg.UniqueTTL)
	}
}

func TestWithProcessIn(t *testing.T) {
	cfg := defaultEnqueueConfig()
	duration := 1 * time.Hour
	opt := WithProcessIn(duration)
	opt(cfg)

	if cfg.ProcessIn != duration {
		t.Errorf("expected ProcessIn %v, got %v", duration, cfg.ProcessIn)
	}
}

func TestWithProcessAt(t *testing.T) {
	cfg := defaultEnqueueConfig()
	now := time.Now().Add(1 * time.Hour)
	opt := WithProcessAt(now)
	opt(cfg)

	if !cfg.ProcessAt.Equal(now) {
		t.Errorf("expected ProcessAt %v, got %v", now, cfg.ProcessAt)
	}
}

func TestWithProcessAtZero(t *testing.T) {
	cfg := defaultEnqueueConfig()
	opt := WithProcessAt(time.Time{})
	opt(cfg)

	if !cfg.ProcessAt.IsZero() {
		t.Errorf("expected ProcessAt zero, got %v", cfg.ProcessAt)
	}
}

func TestWithMuxLogger(t *testing.T) {
	cfg := &muxConfig{}
	logger := zap.NewNop().Sugar()
	opt := WithMuxLogger(logger)
	opt(cfg)

	if cfg.logger != logger {
		t.Error("expected logger to be set")
	}
}
