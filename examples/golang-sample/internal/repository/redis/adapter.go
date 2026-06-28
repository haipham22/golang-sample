package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisClientAdapter adapts a redis.UniversalClient to the narrow cacheClient
// port. It exists so the cache repo depends only on the three commands it
// uses, keeping the surface small and the unit tests free of go-redis
// internals.
type redisClientAdapter struct {
	c redis.UniversalClient
}

// Compile-time guard.
var _ cacheClient = (*redisClientAdapter)(nil)

// adaptRedis wraps a redis.UniversalClient as a cacheClient.
func adaptRedis(c redis.UniversalClient) cacheClient {
	return &redisClientAdapter{c: c}
}

// Get delegates to UniversalClient.Get and translates the *redis.StringCmd
// result/error into a plain (string, error).
func (a *redisClientAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.c.Get(ctx, key).Result()
}

// Set delegates to UniversalClient.Set.
func (a *redisClientAdapter) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return a.c.Set(ctx, key, value, ttl).Err()
}

// Del delegates to UniversalClient.Del and returns the number of keys removed.
func (a *redisClientAdapter) Del(ctx context.Context, key string) (int64, error) {
	return a.c.Del(ctx, key).Result()
}
