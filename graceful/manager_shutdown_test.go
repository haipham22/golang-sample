package graceful

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManagerGoWithError(t *testing.T) {
	m := NewManager(context.TODO(), WithFailFast(true))
	testErr := errors.New("test error")

	started := make(chan struct{})
	m.Go(func(ctx context.Context) error {
		close(started)
		return testErr
	})

	<-started
	// Wait for error to be sent to errCh
	time.Sleep(100 * time.Millisecond)

	err := m.Wait()
	if err != testErr {
		t.Errorf("Wait() = %v, want %v", err, testErr)
	}
}

func TestManagerShutdown(t *testing.T) {
	m := NewManager(context.TODO())
	m.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	time.Sleep(10 * time.Millisecond)
	m.InitiateShutdown()

	err := m.Shutdown(time.Second)
	if err != nil {
		t.Errorf("Shutdown() = %v, want nil", err)
	}
}

func TestManagerShutdownWithTimeout(t *testing.T) {
	m := NewManager(context.TODO())
	m.Go(func(ctx context.Context) error {
		<-ctx.Done()
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	time.Sleep(10 * time.Millisecond)
	m.InitiateShutdown()

	err := m.Shutdown(200 * time.Millisecond)
	if err != nil {
		t.Errorf("Shutdown() = %v, want nil", err)
	}
}

func TestManagerShutdownTimeoutExceeded(t *testing.T) {
	m := NewManager(context.TODO())
	block := make(chan struct{})

	m.Go(func(ctx context.Context) error {
		<-ctx.Done()
		<-block
		return nil
	})

	time.Sleep(10 * time.Millisecond)
	m.InitiateShutdown()

	err := m.Shutdown(50 * time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Errorf("Shutdown() = %v, want %v", err, context.DeadlineExceeded)
	}
	close(block)
}

func TestManagerCleanupError(t *testing.T) {
	m := NewManager(context.TODO())
	testErr := errors.New("cleanup error")

	m.Defer(func(ctx context.Context) error {
		return testErr
	})

	m.InitiateShutdown()
	err := m.Shutdown(time.Second)
	if err != testErr {
		t.Errorf("Shutdown() = %v, want %v", err, testErr)
	}
}
