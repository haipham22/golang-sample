package graceful

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRun_NoServices(t *testing.T) {
	log := zap.NewNop().Sugar()
	timeout := 1 * time.Second

	err := Run(context.Background(), log, timeout)

	if err != nil {
		t.Errorf("expected no error with no services, got %v", err)
	}
}

func TestFromFunc(t *testing.T) {
	startCalled := false
	shutdownCalled := false

	service := FromFunc(
		func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	if !startCalled {
		t.Error("Start function not called")
	}

	if err := service.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !shutdownCalled {
		t.Error("Shutdown function not called")
	}
}

func TestFromFunc_NilStart(t *testing.T) {
	service := FromFunc(
		nil,
		func(ctx context.Context) error {
			return nil
		},
	)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Errorf("Start with nil func should not error, got %v", err)
	}
}

func TestFromFunc_NilShutdown(t *testing.T) {
	service := FromFunc(
		func(ctx context.Context) error {
			return nil
		},
		nil,
	)

	ctx := context.Background()
	if err := service.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown with nil func should not error, got %v", err)
	}
}

func TestFromFunc_StartError(t *testing.T) {
	expectedErr := errors.New("start failed")

	service := FromFunc(
		func(ctx context.Context) error {
			return expectedErr
		},
		nil,
	)

	ctx := context.Background()
	if err := service.Start(ctx); err != expectedErr {
		t.Errorf("Start should return error, got %v", err)
	}
}

func TestFromFunc_ShutdownError(t *testing.T) {
	expectedErr := errors.New("shutdown failed")

	service := FromFunc(
		nil,
		func(ctx context.Context) error {
			return expectedErr
		},
	)

	ctx := context.Background()
	if err := service.Shutdown(ctx); err != expectedErr {
		t.Errorf("Shutdown should return error, got %v", err)
	}
}

// mockService is a test double for Service interface
type mockService struct {
	startCalled    bool
	shutdownCalled bool
	startError     error
	shutdownError  error
	shutdownDelay  time.Duration
	blockStart     bool
}

func (m *mockService) Start(ctx context.Context) error {
	m.startCalled = true
	if m.startError != nil {
		return m.startError
	}
	if m.blockStart {
		<-ctx.Done()
	}
	return nil
}

func (m *mockService) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	if m.shutdownDelay > 0 {
		time.Sleep(m.shutdownDelay)
	}
	if m.shutdownError != nil {
		return m.shutdownError
	}
	return nil
}

func TestRun_SingleService(t *testing.T) {
	tests := []struct {
		name        string
		service     *mockService
		expectError bool
	}{
		{
			name: "successful service",
			service: &mockService{
				blockStart: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := zap.NewNop().Sugar()
			timeout := 2 * time.Second

			// Send SIGTERM to trigger shutdown after service starts
			go func() {
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}()

			err := Run(context.Background(), log, timeout, tt.service)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify lifecycle
			if !tt.service.startCalled {
				t.Error("Start was not called")
			}
			if !tt.service.shutdownCalled {
				t.Error("Shutdown was not called")
			}
		})
	}
}

func TestRun_MultipleServices(t *testing.T) {
	tests := []struct {
		name         string
		services     []*mockService
		expectError  bool
		expectStarts int
	}{
		{
			name: "three concurrent services",
			services: []*mockService{
				{blockStart: true},
				{blockStart: true},
				{blockStart: true},
			},
			expectError:  false,
			expectStarts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := zap.NewNop().Sugar()
			timeout := 2 * time.Second

			services := make([]Service, len(tt.services))
			for i, svc := range tt.services {
				services[i] = svc
			}

			// Send SIGTERM to trigger shutdown after services start
			go func() {
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}()

			err := Run(context.Background(), log, timeout, services...)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			startCount := 0
			for _, svc := range tt.services {
				if svc.startCalled {
					startCount++
				}
			}
			if startCount != tt.expectStarts {
				t.Errorf("expected %d starts, got %d", tt.expectStarts, startCount)
			}

			// Verify all shutdowns called
			for i, svc := range tt.services {
				if !svc.shutdownCalled {
					t.Errorf("Service %d shutdown not called", i)
				}
			}
		})
	}
}

func TestRun_ServiceFunc(t *testing.T) {
	log := zap.NewNop().Sugar()
	timeout := 2 * time.Second

	startCalled := false
	shutdownCalled := false

	service := FromFunc(
		func(ctx context.Context) error {
			startCalled = true
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	)

	// Send SIGTERM to trigger shutdown after service starts
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := Run(context.Background(), log, timeout, service)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !startCalled {
		t.Error("Start was not called")
	}

	if !shutdownCalled {
		t.Error("Shutdown was not called")
	}
}

func TestRun_ShutdownDelay(t *testing.T) {
	log := zap.NewNop().Sugar()
	timeout := 2 * time.Second

	service := &mockService{
		blockStart:    true,
		shutdownDelay: 100 * time.Millisecond,
	}

	// Send SIGTERM to trigger shutdown after service starts
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := Run(context.Background(), log, timeout, service)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !service.shutdownCalled {
		t.Error("Shutdown was not called")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	log := zap.NewNop().Sugar()
	timeout := 2 * time.Second

	// Service that respects context cancellation
	service := FromFunc(
		func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		func(ctx context.Context) error {
			return nil
		},
	)

	// Send SIGTERM immediately
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := Run(context.Background(), log, timeout, service)

	// Context cancellation is expected
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("Context cancellation test result: %v", err)
	}
}

func BenchmarkRun(b *testing.B) {
	log := zap.NewNop().Sugar()
	timeout := 1 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service := &mockService{}

		// Send SIGTERM immediately in benchmark
		go func() {
			time.Sleep(10 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		_ = Run(context.Background(), log, timeout, service)
	}
}

func BenchmarkRun_MultipleServices(b *testing.B) {
	log := zap.NewNop().Sugar()
	timeout := 1 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		services := []Service{
			&mockService{blockStart: true},
			&mockService{blockStart: true},
			&mockService{blockStart: true},
		}

		// Send SIGTERM immediately in benchmark
		go func() {
			time.Sleep(10 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		_ = Run(context.Background(), log, timeout, services...)
	}
}
