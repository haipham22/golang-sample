package graceful

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerGroup(t *testing.T) {
	wg := NewWorkerGroup(10)
	if wg == nil {
		t.Fatal("NewWorkerGroup() returned nil")
	}
	if wg.sem == nil {
		t.Error("sem channel is nil")
	}
}

func TestNewWorkerGroupZeroConcurrency(t *testing.T) {
	wg := NewWorkerGroup(0)
	if wg == nil {
		t.Fatal("NewWorkerGroup() returned nil")
	}
}

func TestWorkerGroupTryGo(t *testing.T) {
	wg := NewWorkerGroup(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	if !wg.TryGo(ctx, func(ctx context.Context) {
		close(done)
	}) {
		t.Fatal("TryGo() returned false")
	}

	select {
	case <-done:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("TryGo() did not execute function")
	}
}

func TestWorkerGroupTryGoCanceled(t *testing.T) {
	wg := NewWorkerGroup(2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if wg.TryGo(ctx, func(ctx context.Context) {}) {
		t.Error("TryGo() should return false when ctx is canceled")
	}
}

func TestWorkerGroupConcurrency(t *testing.T) {
	wg := NewWorkerGroup(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int64
	var maxCount atomic.Int64

	// Start 5 jobs, but max concurrency is 2
	for i := 0; i < 5; i++ {
		wg.TryGo(ctx, func(ctx context.Context) {
			c := count.Add(1)
			for {
				m := maxCount.Load()
				if c <= m || maxCount.CompareAndSwap(m, c) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			count.Add(-1)
		})
	}

	if maxCount.Load() > 2 {
		t.Errorf("Max concurrency = %v, want <= 2", maxCount.Load())
	}
}

func TestWorkerGroupDrain(t *testing.T) {
	wg := NewWorkerGroup(2)
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < 3; i++ {
		wg.TryGo(ctx, func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})
	}

	cancel()
	err := wg.Drain(time.Second)
	if err != nil {
		t.Errorf("Drain() = %v, want nil", err)
	}
}

func TestWorkerGroupDrainTimeout(t *testing.T) {
	wg := NewWorkerGroup(1)
	ctx, cancel := context.WithCancel(context.Background())

	// Start a job that never finishes
	wg.TryGo(ctx, func(ctx context.Context) {
		<-ctx.Done()
		time.Sleep(100 * time.Millisecond)
	})

	cancel()

	err := wg.Drain(10 * time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Errorf("Drain() = %v, want %v", err, context.DeadlineExceeded)
	}
}
