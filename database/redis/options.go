package redis

import "time"

type Option func(*Config)

type Config struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	MaxRetryBackoff time.Duration
	MinRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ConnMaxIdleTime time.Duration

	// Cluster support for UniversalClient
	Addrs          []string // Cluster addresses
	RouteByLatency bool     // Route by latency
	RouteRandomly  bool     // Random route selection
}

func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Addr = addr
	}
}

func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

func WithDB(db int) Option {
	return func(c *Config) {
		c.DB = db
	}
}

func WithPoolSize(n int) Option {
	return func(c *Config) {
		c.PoolSize = n
	}
}

func WithMinIdleConns(n int) Option {
	return func(c *Config) {
		c.MinIdleConns = n
	}
}

func WithMaxRetries(n int) Option {
	return func(c *Config) {
		c.MaxRetries = n
	}
}

func WithMaxRetryBackoff(d time.Duration) Option {
	return func(c *Config) {
		c.MaxRetryBackoff = d
	}
}

func WithMinRetryBackoff(d time.Duration) Option {
	return func(c *Config) {
		c.MinRetryBackoff = d
	}
}

func WithDialTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.DialTimeout = d
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = d
	}
}

func WithPoolTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.PoolTimeout = d
	}
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = d
	}
}

// WithAddrs sets Redis cluster addresses.
// For single-node, use WithAddr or pass a single address.
func WithAddrs(addrs ...string) Option {
	return func(c *Config) {
		c.Addrs = addrs
	}
}

// WithRouteByLatency enables latency-based routing for cluster.
func WithRouteByLatency(v bool) Option {
	return func(c *Config) {
		c.RouteByLatency = v
	}
}

// WithRouteRandomly enables random shard selection for cluster.
func WithRouteRandomly(v bool) Option {
	return func(c *Config) {
		c.RouteRandomly = v
	}
}

func defaultConfig() *Config {
	return &Config{
		PoolSize:        DefaultPoolSize,
		MinIdleConns:    DefaultMinIdle,
		MaxRetries:      3,
		MaxRetryBackoff: 500 * time.Millisecond,
		MinRetryBackoff: 8 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}
