package graceful

import "context"

// Service represents a service with lifecycle management
type Service interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// ServiceFunc converts functions into Service
type ServiceFunc struct {
	StartFunc    func(ctx context.Context) error
	ShutdownFunc func(ctx context.Context) error
}

func (s *ServiceFunc) Start(ctx context.Context) error {
	if s.StartFunc == nil {
		return nil
	}
	return s.StartFunc(ctx)
}

func (s *ServiceFunc) Shutdown(ctx context.Context) error {
	if s.ShutdownFunc == nil {
		return nil
	}
	return s.ShutdownFunc(ctx)
}

// FromFunc creates Service from start/shutdown functions
func FromFunc(
	start func(ctx context.Context) error,
	shutdown func(ctx context.Context) error,
) Service {
	return &ServiceFunc{
		StartFunc:    start,
		ShutdownFunc: shutdown,
	}
}
