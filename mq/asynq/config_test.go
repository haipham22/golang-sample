package asynq

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Concurrency != DefaultConcurrency {
		t.Errorf("expected Concurrency %d, got %d", DefaultConcurrency, cfg.Concurrency)
	}

	if cfg.Queues[DefaultQueue] != 1 {
		t.Errorf("expected default queue priority 1, got %d", cfg.Queues[DefaultQueue])
	}

	if cfg.ShutdownTimeout != DefaultShutdownTimeout {
		t.Errorf("expected ShutdownTimeout %v, got %v", DefaultShutdownTimeout, cfg.ShutdownTimeout)
	}
}

func TestDefaultClientConfig(t *testing.T) {
	cfg := defaultClientConfig()

	if cfg.MaxRetry != DefaultMaxRetry {
		t.Errorf("expected MaxRetry %d, got %d", DefaultMaxRetry, cfg.MaxRetry)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, cfg.Timeout)
	}
}

func TestDefaultEnqueueConfig(t *testing.T) {
	cfg := defaultEnqueueConfig()

	if cfg.Queue != DefaultQueue {
		t.Errorf("expected Queue %s, got %s", DefaultQueue, cfg.Queue)
	}

	if cfg.MaxRetry != DefaultMaxRetry {
		t.Errorf("expected MaxRetry %d, got %d", DefaultMaxRetry, cfg.MaxRetry)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, cfg.Timeout)
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"DefaultConcurrency", DefaultConcurrency, 10},
		{"DefaultQueue", DefaultQueue, "default"},
		{"DefaultMaxRetry", DefaultMaxRetry, 25},
		{"DefaultTimeout", DefaultTimeout, 30 * time.Minute},
		{"DefaultShutdownTimeout", DefaultShutdownTimeout, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.value)
			}
		})
	}
}
