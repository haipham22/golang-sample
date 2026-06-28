// Package redis contains Redis-backed repositories. The cache Repository
// wraps a redis.UniversalClient (as returned by govern/database/redis) behind a
// narrow, unexported cacheClient port so the cache logic can be unit tested
// without a live Redis instance (the fake implements only the 3 commands used).
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// Repository is the cache persistence contract.
type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// cacheClient is the narrow subset of redis.UniversalClient the cache repo
// needs. redisClientAdapter adapts a real *redis.Client/UniversalClient to it.
// A hand-written fake in cache_test.go implements it directly.
type cacheClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) (int64, error)
}

// repo is the Redis-backed Repository implementation.
type repo struct {
	log    *zap.SugaredLogger
	client cacheClient
}

// Compile-time guard.
var _ Repository = (*repo)(nil)

// New wires a cache Repository from a govern/database/redis client (or any
// redis.UniversalClient). Mirrors repository/user.New(log, db).
func New(log *zap.SugaredLogger, client redis.UniversalClient) Repository {
	return &repo{log: log, client: adaptRedis(client)}
}

// newWithClient is an unexported constructor used by tests to inject a fake
// cacheClient directly.
func newWithClient(log *zap.SugaredLogger, client cacheClient) Repository {
	return &repo{log: log, client: client}
}

// Get returns the cached value for key. A missing key yields an
// apperrors.NotFound so callers can distinguish "absent" from "hard error"
// without inspecting redis.Nil directly.
func (r *repo) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", apperrors.NewCode(apperrors.CodeInvalid, "cache key is required")
	}
	val, err := r.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil { // go-redis sentinel for missing keys
			return "", apperrors.NotFound("cache key " + key)
		}
		r.log.Errorf("cache get failed for key %q: %v", key, err)
		return "", apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return val, nil
}

// Set stores value under key with the given TTL.
func (r *repo) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if key == "" {
		return apperrors.NewCode(apperrors.CodeInvalid, "cache key is required")
	}
	if err := r.client.Set(ctx, key, value, ttl); err != nil {
		r.log.Errorf("cache set failed for key %q: %v", key, err)
		return apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return nil
}

// Delete removes key. Missing keys are not an error.
func (r *repo) Delete(ctx context.Context, key string) error {
	if key == "" {
		return apperrors.NewCode(apperrors.CodeInvalid, "cache key is required")
	}
	if _, err := r.client.Del(ctx, key); err != nil {
		r.log.Errorf("cache delete failed for key %q: %v", key, err)
		return apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return nil
}
