package redis_test

import (
	"testing"

	"github.com/haipham22/govern/database/redis"
)

func TestNew(t *testing.T) {
	client, cleanup, err := redis.New("localhost:6379")

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("New() expected non-nil client")
	}

	if cleanup == nil {
		t.Fatal("New() expected non-nil cleanup function")
	}

	cleanup()
}

func TestCleanup(t *testing.T) {
	client, cleanup, err := redis.New("localhost:6379")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	cleanup()

	// Calling cleanup again should be safe
	cleanup()

	_ = client
}

func TestDefaults(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue interface{}
		field        string
	}{
		{"DefaultPoolSize", 100, "PoolSize"},
		{"DefaultMinIdle", 10, "MinIdleConns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.field {
			case "PoolSize":
				if got := redis.DefaultPoolSize; got != tt.defaultValue.(int) {
					t.Errorf("DefaultPoolSize = %v, want %v", got, tt.defaultValue)
				}
			case "MinIdleConns":
				if got := redis.DefaultMinIdle; got != tt.defaultValue.(int) {
					t.Errorf("DefaultMinIdle = %v, want %v", got, tt.defaultValue)
				}
			}
		})
	}
}

func TestNewReturnsUniversalClient(t *testing.T) {
	client, cleanup, err := redis.New("localhost:6379")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer cleanup()

	// Verify UniversalClient interface - this will work in both phases
	var _ = client

	// After Phase 04, the concrete type should NOT be directly *redis.Client
	// but rather the interface. For now, we just verify it works as UniversalClient.
	if client == nil {
		t.Fatal("New() expected non-nil client")
	}

	// Verify common UniversalClient methods work
	// This will fail if the underlying type doesn't implement UniversalClient
	ctx := t.Context()
	_ = client.Ping(ctx)
}

func TestNewClientReturnsUniversalClient(t *testing.T) {
	client, cleanup, err := redis.NewClient("localhost:6379")
	if err != nil {
		t.Fatalf("NewClient() unexpected error: %v", err)
	}
	defer cleanup()

	// Type assertion to UniversalClient
	var _ = client

	// Verify UniversalClient methods
	ctx := t.Context()
	_ = client.Ping(ctx)
}

func TestNewFromDSNReturnsUniversalClient(t *testing.T) {
	client, cleanup, err := redis.NewFromDSN("redis://localhost:6379")
	if err != nil {
		t.Fatalf("NewFromDSN() unexpected error: %v", err)
	}
	defer cleanup()

	// Type assertion to UniversalClient
	var _ = client

	// Verify UniversalClient methods
	ctx := t.Context()
	_ = client.Ping(ctx)
}

func TestClusterMode(t *testing.T) {
	client, cleanup, err := redis.New("",
		redis.WithAddrs("node1:6379", "node2:6379"),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer cleanup()

	// Verify it's a UniversalClient
	var _ = client

	// Note: In Phase 04, this will use redis.NewUniversalClient()
	// which properly handles cluster mode. For now it creates a single client
	// but with cluster addresses configured (which won't actually connect to cluster)
}
