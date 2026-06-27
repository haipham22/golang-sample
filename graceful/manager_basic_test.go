package graceful

import (
	"context"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager(context.TODO())
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.ctx == nil {
		t.Error("ctx is nil")
	}
	if m.failFast != true {
		t.Error("default failFast should be true")
	}
}

func TestNewManagerWithParent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NewManager(ctx)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestNewManagerWithOptions(t *testing.T) {
	m := NewManager(context.TODO(), WithFailFast(false))
	if m.failFast != false {
		t.Error("WithFailFast(false) failed")
	}
}

func TestManagerContext(t *testing.T) {
	m := NewManager(context.TODO())
	ctx := m.Context()
	if ctx == nil {
		t.Fatal("Context() returned nil")
	}
}

func TestManagerInitiateShutdown(t *testing.T) {
	m := NewManager(context.TODO())

	// Should not panic
	m.InitiateShutdown()
	m.InitiateShutdown() // Idempotent

	select {
	case <-m.Context().Done():
		// Expected
	default:
		t.Error("Context should be canceled after InitiateShutdown")
	}
}

func TestManagerGo(t *testing.T) {
	m := NewManager(context.TODO())
	done := make(chan struct{})

	m.Go(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		close(done)
		return nil
	})

	select {
	case <-done:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("Go() did not execute function")
	}
}

func TestManagerDefer(t *testing.T) {
	m := NewManager(context.TODO())
	called := false

	m.Defer(func(ctx context.Context) error {
		called = true
		return nil
	})

	m.InitiateShutdown()
	_ = m.Shutdown(1000)

	if !called {
		t.Error("Cleanup was not called")
	}
}

func TestManagerDeferLIFO(t *testing.T) {
	m := NewManager(context.TODO())
	order := []int{}

	m.Defer(func(ctx context.Context) error {
		order = append(order, 1)
		return nil
	})
	m.Defer(func(ctx context.Context) error {
		order = append(order, 2)
		return nil
	})
	m.Defer(func(ctx context.Context) error {
		order = append(order, 3)
		return nil
	})

	m.InitiateShutdown()
	_ = m.Shutdown(1000)

	if len(order) != 3 || order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("Cleanups did not run in LIFO order: %v", order)
	}
}

func TestManagerWait(t *testing.T) {
	m := NewManager(context.TODO())
	m.InitiateShutdown()

	err := m.Wait()
	if err != nil {
		t.Errorf("Wait() = %v, want nil", err)
	}
}
