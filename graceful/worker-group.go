package graceful

import (
	"context"
	"sync"
	"time"
)

// WorkerGroup runs jobs with concurrency, stops intake on ctx cancel,
// and waits for inflight work to finish (bounded by timeout).
type WorkerGroup struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

// NewWorkerGroup creates a new WorkerGroup with max concurrency.
// If concurrency <= 0, defaults to 1.
func NewWorkerGroup(concurrency int) *WorkerGroup {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &WorkerGroup{sem: make(chan struct{}, concurrency)}
}

// TryGo starts a unit of work if ctx not canceled.
// Returns false if shutdown has started and you should stop intake.
func (g *WorkerGroup) TryGo(ctx context.Context, fn func(context.Context)) bool {
	// stop intake if shutting down
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// acquire slot or stop
	select {
	case g.sem <- struct{}{}:
	case <-ctx.Done():
		return false
	}

	g.wg.Add(1)
	go func() {
		defer func() {
			<-g.sem
			g.wg.Done()
		}()
		fn(ctx)
	}()
	return true
}

// Drain waits for inflight work to finish, bounded by timeout.
// Returns context.DeadlineExceeded if timeout is reached before all work completes.
func (g *WorkerGroup) Drain(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	if timeout <= 0 {
		<-done
		return nil
	}
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}
