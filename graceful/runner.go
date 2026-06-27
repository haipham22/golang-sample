package graceful

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Run starts all services and handles graceful shutdown
func Run(
	ctx context.Context,
	log *zap.SugaredLogger,
	shutdownTimeout time.Duration,
	services ...Service,
) error {
	if len(services) == 0 {
		log.Warn("No services provided, nothing to run")
		return nil
	}

	// Create graceful shutdown manager
	m := NewManager(ctx, WithLogger(log))

	// Register all service shutdown hooks
	for _, svc := range services {
		m.Defer(svc.Shutdown)
	}

	// Start all services in goroutines
	for i, svc := range services {
		idx := i
		m.Go(func(ctx context.Context) error {
			log.Infof("Starting service %d", idx)
			return svc.Start(ctx)
		})
	}

	log.Info("All services started")

	// Wait for shutdown signal
	if err := m.Wait(); err != nil {
		log.Errorf("Error waiting for shutdown: %v", err)
		return err
	}

	// Perform graceful shutdown
	log.Info("Shutting down...")
	if err := m.Shutdown(shutdownTimeout); err != nil {
		log.Errorf("Shutdown error: %v", err)
		return err
	}

	log.Info("Graceful shutdown complete")
	return nil
}
