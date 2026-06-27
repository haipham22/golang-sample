package redis_test

import (
	"testing"
	"time"

	"github.com/haipham22/govern/database/redis"
)

func TestClusterOptions(t *testing.T) {
	tests := []struct {
		name   string
		option redis.Option
		verify func(*redis.Config) bool
	}{
		{
			name:   "WithAddrs sets multiple addresses",
			option: redis.WithAddrs("node1:6379", "node2:6379", "node3:6379"),
			verify: func(cfg *redis.Config) bool {
				return len(cfg.Addrs) == 3 &&
					cfg.Addrs[0] == "node1:6379" &&
					cfg.Addrs[1] == "node2:6379" &&
					cfg.Addrs[2] == "node3:6379"
			},
		},
		{
			name:   "WithAddrs sets single address",
			option: redis.WithAddrs("localhost:6379"),
			verify: func(cfg *redis.Config) bool {
				return len(cfg.Addrs) == 1 && cfg.Addrs[0] == "localhost:6379"
			},
		},
		{
			name:   "WithRouteByLatency true",
			option: redis.WithRouteByLatency(true),
			verify: func(cfg *redis.Config) bool {
				return cfg.RouteByLatency == true
			},
		},
		{
			name:   "WithRouteByLatency false",
			option: redis.WithRouteByLatency(false),
			verify: func(cfg *redis.Config) bool {
				return cfg.RouteByLatency == false
			},
		},
		{
			name:   "WithRouteRandomly true",
			option: redis.WithRouteRandomly(true),
			verify: func(cfg *redis.Config) bool {
				return cfg.RouteRandomly == true
			},
		},
		{
			name:   "WithRouteRandomly false",
			option: redis.WithRouteRandomly(false),
			verify: func(cfg *redis.Config) bool {
				return cfg.RouteRandomly == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &redis.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option did not set expected value")
			}
		})
	}
}

func TestTimeoutOptions(t *testing.T) {
	tests := []struct {
		name   string
		option redis.Option
		verify func(*redis.Config) bool
	}{
		{
			name:   "WithDialTimeout",
			option: redis.WithDialTimeout(5 * time.Second),
			verify: func(cfg *redis.Config) bool {
				return cfg.DialTimeout == 5*time.Second
			},
		},
		{
			name:   "WithReadTimeout",
			option: redis.WithReadTimeout(3 * time.Second),
			verify: func(cfg *redis.Config) bool {
				return cfg.ReadTimeout == 3*time.Second
			},
		},
		{
			name:   "WithWriteTimeout",
			option: redis.WithWriteTimeout(3 * time.Second),
			verify: func(cfg *redis.Config) bool {
				return cfg.WriteTimeout == 3*time.Second
			},
		},
		{
			name:   "WithPoolTimeout",
			option: redis.WithPoolTimeout(4 * time.Second),
			verify: func(cfg *redis.Config) bool {
				return cfg.PoolTimeout == 4*time.Second
			},
		},
		{
			name:   "WithConnMaxIdleTime",
			option: redis.WithConnMaxIdleTime(5 * time.Minute),
			verify: func(cfg *redis.Config) bool {
				return cfg.ConnMaxIdleTime == 5*time.Minute
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &redis.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option did not set expected value")
			}
		})
	}
}

func TestRetryOptions(t *testing.T) {
	tests := []struct {
		name   string
		option redis.Option
		verify func(*redis.Config) bool
	}{
		{
			name:   "WithMaxRetries",
			option: redis.WithMaxRetries(5),
			verify: func(cfg *redis.Config) bool {
				return cfg.MaxRetries == 5
			},
		},
		{
			name:   "WithMaxRetryBackoff",
			option: redis.WithMaxRetryBackoff(500 * time.Millisecond),
			verify: func(cfg *redis.Config) bool {
				return cfg.MaxRetryBackoff == 500*time.Millisecond
			},
		},
		{
			name:   "WithMinRetryBackoff",
			option: redis.WithMinRetryBackoff(8 * time.Millisecond),
			verify: func(cfg *redis.Config) bool {
				return cfg.MinRetryBackoff == 8*time.Millisecond
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &redis.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option did not set expected value")
			}
		})
	}
}

func TestAddrOption(t *testing.T) {
	t.Run("WithAddr sets address", func(t *testing.T) {
		cfg := &redis.Config{}
		redis.WithAddr("custom.redis.com:6380")(cfg)

		if cfg.Addr != "custom.redis.com:6380" {
			t.Errorf("WithAddr() = %v, want %v", cfg.Addr, "custom.redis.com:6380")
		}
	})
}

func TestOptionValidation(t *testing.T) {
	tests := []struct {
		name   string
		option redis.Option
		verify func(*redis.Config) bool
	}{
		{
			name:   "zero pool size",
			option: redis.WithPoolSize(0),
			verify: func(cfg *redis.Config) bool {
				return cfg.PoolSize == 0
			},
		},
		{
			name:   "negative pool size",
			option: redis.WithPoolSize(-10),
			verify: func(cfg *redis.Config) bool {
				return cfg.PoolSize == -10
			},
		},
		{
			name:   "zero min idle conns",
			option: redis.WithMinIdleConns(0),
			verify: func(cfg *redis.Config) bool {
				return cfg.MinIdleConns == 0
			},
		},
		{
			name:   "negative min idle conns",
			option: redis.WithMinIdleConns(-5),
			verify: func(cfg *redis.Config) bool {
				return cfg.MinIdleConns == -5
			},
		},
		{
			name:   "zero timeout",
			option: redis.WithDialTimeout(0),
			verify: func(cfg *redis.Config) bool {
				return cfg.DialTimeout == 0
			},
		},
		{
			name:   "negative db number",
			option: redis.WithDB(-1),
			verify: func(cfg *redis.Config) bool {
				return cfg.DB == -1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &redis.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option validation failed")
			}
		})
	}
}
