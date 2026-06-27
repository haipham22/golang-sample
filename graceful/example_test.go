package graceful

import (
	"context"
	"net/http"
	"time"
)

// ExampleManager demonstrates basic Manager usage.
func ExampleManager() {
	m := NewManager(context.TODO())

	// Run a background goroutine
	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				// Do work
				time.Sleep(100 * time.Millisecond)
			}
		}
	})

	// Register cleanup
	m.Defer(func(ctx context.Context) error {
		// Cleanup resources
		return nil
	})

	// Wait for signal
	_ = m.Wait()
	_ = m.Shutdown(10 * time.Second)
}

// ExampleManager_withFailFast demonstrates fail-fast behavior.
func ExampleManager_withFailFast() {
	m := NewManager(context.TODO(), WithFailFast(true))

	m.Go(func(ctx context.Context) error {
		// If this returns an error, shutdown begins immediately
		return context.Canceled
	})

	_ = m.Wait()
	_ = m.Shutdown(5 * time.Second)
}

// ExampleWorkerGroup demonstrates WorkerGroup usage.
func ExampleWorkerGroup() {
	m := NewManager(context.TODO())
	wg := NewWorkerGroup(10) // Max 10 concurrent jobs

	m.Go(func(ctx context.Context) error {
		for i := 0; i < 100; i++ {
			if !wg.TryGo(ctx, func(ctx context.Context) {
				// Process job
			}) {
				break
			}
		}
		return nil
	})

	// Drain waits for all inflight jobs
	m.Defer(func(ctx context.Context) error {
		return wg.Drain(5 * time.Second)
	})

	_ = m.Wait()
	_ = m.Shutdown(10 * time.Second)
}

// ExampleServer demonstrates HTTP server usage.
// Note: HTTP server functionality has moved to github.com/haipham22/govern/http
func ExampleServer() {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	})

	// Import github.com/haipham22/govern/http
	// server := http.NewServer(
	//     ":8080",
	//     handler,
	//     http.WithShutdownTimeout(10*time.Second),
	// )
	// _ = server.Start()
}
