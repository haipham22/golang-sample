package redis

import "github.com/redis/go-redis/v9"

const (
	DefaultPoolSize = 100
	DefaultMinIdle  = 10
)

// New creates a new Redis client.
// Returns UniversalClient which supports both standalone and cluster modes.
func New(addr string, opts ...Option) (redis.UniversalClient, func(), error) {
	return NewClient(addr, opts...)
}

// NewClient creates a new Redis client with an explicit address.
// Returns UniversalClient interface supporting both single-node and cluster configurations.
//
// For single-node Redis, pass a single address. For cluster mode, use WithAddrs()
// with multiple node addresses.
//
// The returned cleanup function closes the client connection and should be
// called when the client is no longer needed (typically via defer).
func NewClient(addr string, opts ...Option) (redis.UniversalClient, func(), error) {
	cfg := defaultConfig()

	// Handle single addr -> convert to Addrs slice for UniversalClient
	if addr != "" {
		cfg.Addrs = []string{addr}
	}
	cfg.Addr = addr // Keep for reference

	for _, opt := range opts {
		opt(cfg)
	}

	// Fallback: if Addrs empty but Addr set, use Addr
	if len(cfg.Addrs) == 0 && cfg.Addr != "" {
		cfg.Addrs = []string{cfg.Addr}
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:           cfg.Addrs,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		MinRetryBackoff: cfg.MinRetryBackoff,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		RouteByLatency:  cfg.RouteByLatency,
		RouteRandomly:   cfg.RouteRandomly,
	})

	cleanup := func() {
		_ = client.Close()
	}

	return client, cleanup, nil
}

// NewFromDSN creates a new Redis client from a DSN string.
// DSN format: redis://[:password@]host[:port][/db][?options]
// Returns UniversalClient interface.
func NewFromDSN(dsn string, opts ...Option) (redis.UniversalClient, func(), error) {
	addr, parsedOpts, err := ParseDSN(dsn)
	if err != nil {
		return nil, nil, err
	}

	// Explicit opts override parsed opts
	parsedOpts = append(parsedOpts, opts...)
	return NewClient(addr, parsedOpts...)
}
