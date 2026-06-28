package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// fakeClient is a hand-written cacheClient double. It does NOT require a live
// Redis (no miniredis dependency) and lets each test dial in the desired
// response per operation.
type fakeClient struct {
	data        map[string]string
	getErr      error // when set, Get returns this error (e.g. redis.Nil)
	setErr      error
	delErr      error
	lastSetTTL  time.Duration
	lastSetKey  string
	lastSetValue string
}

func newFakeClient() *fakeClient {
	return &fakeClient{data: make(map[string]string)}
}

func (f *fakeClient) Get(_ context.Context, key string) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	v, ok := f.data[key]
	if !ok {
		return "", redis.Nil
	}
	return v, nil
}

func (f *fakeClient) Set(_ context.Context, key, value string, ttl time.Duration) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.data[key] = value
	f.lastSetKey = key
	f.lastSetValue = value
	f.lastSetTTL = ttl
	return nil
}

func (f *fakeClient) Del(_ context.Context, key string) (int64, error) {
	if f.delErr != nil {
		return 0, f.delErr
	}
	if _, ok := f.data[key]; ok {
		delete(f.data, key)
		return 1, nil
	}
	return 0, nil
}

func TestRepository_Get(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		fc := newFakeClient()
		_ = fc.Set(context.Background(), "k", "v", 0)
		r := newWithClient(zap.NewNop().Sugar(), fc)

		got, err := r.Get(context.Background(), "k")
		require.NoError(t, err)
		assert.Equal(t, "v", got)
	})

	t.Run("miss -> NotFound", func(t *testing.T) {
		r := newWithClient(zap.NewNop().Sugar(), newFakeClient())
		_, err := r.Get(context.Background(), "missing")
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})

	t.Run("hard error -> Internal", func(t *testing.T) {
		fc := newFakeClient()
		fc.getErr = errors.New("conn reset")
		r := newWithClient(zap.NewNop().Sugar(), fc)

		_, err := r.Get(context.Background(), "k")
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
	})

	t.Run("empty key rejected", func(t *testing.T) {
		r := newWithClient(zap.NewNop().Sugar(), newFakeClient())
		_, err := r.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})
}

func TestRepository_Set(t *testing.T) {
	t.Run("ok with ttl", func(t *testing.T) {
		fc := newFakeClient()
		r := newWithClient(zap.NewNop().Sugar(), fc)

		require.NoError(t, r.Set(context.Background(), "k", "v", 5*time.Second))
		assert.Equal(t, "v", fc.data["k"])
		assert.Equal(t, 5*time.Second, fc.lastSetTTL)
	})

	t.Run("empty key rejected", func(t *testing.T) {
		r := newWithClient(zap.NewNop().Sugar(), newFakeClient())
		err := r.Set(context.Background(), "", "v", 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("error propagated", func(t *testing.T) {
		fc := newFakeClient()
		fc.setErr = errors.New("write failed")
		r := newWithClient(zap.NewNop().Sugar(), fc)

		err := r.Set(context.Background(), "k", "v", 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		fc := newFakeClient()
		_ = fc.Set(context.Background(), "k", "v", 0)
		r := newWithClient(zap.NewNop().Sugar(), fc)

		require.NoError(t, r.Delete(context.Background(), "k"))
		_, ok := fc.data["k"]
		assert.False(t, ok)
	})

	t.Run("missing key is not an error", func(t *testing.T) {
		r := newWithClient(zap.NewNop().Sugar(), newFakeClient())
		require.NoError(t, r.Delete(context.Background(), "absent"))
	})

	t.Run("empty key rejected", func(t *testing.T) {
		r := newWithClient(zap.NewNop().Sugar(), newFakeClient())
		err := r.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("error propagated", func(t *testing.T) {
		fc := newFakeClient()
		fc.delErr = errors.New("del failed")
		r := newWithClient(zap.NewNop().Sugar(), fc)

		err := r.Delete(context.Background(), "k")
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
	})
}

func TestRepository_InterfaceCompliance(t *testing.T) {
	var _ Repository = (*repo)(nil)
	var _ cacheClient = (*redisClientAdapter)(nil)
}
