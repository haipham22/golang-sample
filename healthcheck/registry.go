package healthcheck

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type registeredCheck struct {
	fn     Check
	config *checkConfig
}

// Registry manages health checks.
type Registry struct {
	mu     sync.RWMutex
	checks map[string]*registeredCheck
}

// New creates a new health check registry.
func New() *Registry {
	return &Registry{
		checks: make(map[string]*registeredCheck),
	}
}

// Register registers a new health check.
// Panics if a check with the same name already exists.
func (r *Registry) Register(name string, check Check, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.checks[name]; exists {
		panic(fmt.Sprintf("health check already registered: %s", name))
	}

	cfg := &checkConfig{
		timeout:      5 * time.Second,
		disablePanic: false, // enable panic recovery by default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	r.checks[name] = &registeredCheck{
		fn:     check,
		config: cfg,
	}
}

// Unregister removes a health check.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.checks, name)
}

// Run executes all registered health checks.
func (r *Registry) Run(ctx context.Context) *Response {
	start := time.Now()
	resp := &Response{
		Timestamp: start,
		Status:    StatusPassing,
		Checks:    make(map[string]Result),
	}

	r.mu.RLock()
	checkNames := make([]string, 0, len(r.checks))
	for name := range r.checks {
		checkNames = append(checkNames, name)
	}
	r.mu.RUnlock()

	sort.Strings(checkNames)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range checkNames {
		wg.Add(1)
		go func(checkName string) {
			defer wg.Done()

			r.mu.RLock()
			check := r.checks[checkName]
			r.mu.RUnlock()

			result := r.runCheck(ctx, checkName, check)

			mu.Lock()
			resp.Checks[checkName] = result
			mu.Unlock()

			if result.Status == StatusFailing {
				mu.Lock()
				resp.Status = StatusFailing
				mu.Unlock()
			} else if result.Status == StatusWarning && resp.Status == StatusPassing {
				mu.Lock()
				resp.Status = StatusWarning
				mu.Unlock()
			}
		}(name)
	}

	wg.Wait()
	resp.Duration = time.Since(start)
	return resp
}

func (r *Registry) runCheck(ctx context.Context, name string, check *registeredCheck) Result {
	start := time.Now()
	result := Result{
		Name:      name,
		Timestamp: start,
		Status:    StatusPassing,
	}

	checkCtx := ctx
	if check.config.timeout > 0 {
		var cancel context.CancelFunc
		checkCtx, cancel = context.WithTimeout(ctx, check.config.timeout)
		defer cancel()
	}

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		done <- check.fn(checkCtx)
	}()

	select {
	case <-checkCtx.Done():
		result.Status = StatusFailing
		result.Message = "timeout"
	case err := <-done:
		result.Duration = time.Since(start)
		if err != nil {
			result.Status = StatusFailing
			result.Message = err.Error()
		}
	}

	return result
}
