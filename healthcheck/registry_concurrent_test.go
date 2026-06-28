package healthcheck

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_SortedChecks(t *testing.T) {
	r := New()

	checks := []string{"zebra", "apple", "banana", "cherry"}
	for _, name := range checks {
		n := name
		r.Register(n, func(ctx context.Context) error {
			return nil
		})
	}

	resp := r.Run(context.Background())

	// Check that all checks are present
	names := make([]string, 0, len(resp.Checks))
	for name := range resp.Checks {
		names = append(names, name)
	}
	sort.Strings(names)
	assert.Equal(t, []string{"apple", "banana", "cherry", "zebra"}, names)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := New()

	// Concurrent registration
	done := make(chan bool)
	for i := range 10 {
		go func(n int) {
			r.Register(string(rune('a'+n)), func(ctx context.Context) error {
				return nil
			})
			done <- true
		}(i)
	}

	for range 10 {
		<-done
	}

	// Concurrent execution
	for range 5 {
		go func() {
			r.Run(context.Background())
			done <- true
		}()
	}

	for range 5 {
		<-done
	}
}

func TestRegistry_ResponseTimestamp(t *testing.T) {
	r := New()

	r.Register("check", func(ctx context.Context) error {
		return nil
	})

	before := time.Now()
	resp := r.Run(context.Background())
	after := time.Now()

	assert.True(t, resp.Timestamp.After(before.Add(-time.Second)))
	assert.True(t, resp.Timestamp.Before(after.Add(time.Second)))
}

func TestCheckResultTimestamp(t *testing.T) {
	r := New()

	r.Register("check", func(ctx context.Context) error {
		return nil
	})

	before := time.Now()
	resp := r.Run(context.Background())
	after := time.Now()

	result := resp.Checks["check"]
	assert.True(t, result.Timestamp.After(before.Add(-time.Second)))
	assert.True(t, result.Timestamp.Before(after.Add(time.Second)))
}
