// Package grpc contains a minimal gRPC server template. govern does not ship a
// grpc helper, so this package uses google.golang.org/grpc directly.
//
// To keep the template self-contained (no .proto / codegen toolchain), the
// sample Greeter service is implemented against grpc's generic unary-handler
// API. In a real project, generate stubs with protoc-gen-go-grpc and replace
// registerGreeter with the generated RegisterXxxServer call.
package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GreeterServer is the server-side contract for the sample Greeter service.
// Generated services would satisfy a richer interface; this minimal version
// exposes a single SayHello RPC.
type GreeterServer interface {
	SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)
}

// HelloRequest is the request message for Greeter.SayHello.
type HelloRequest struct {
	Name string
}

// HelloReply is the response message for Greeter.SayHello.
type HelloReply struct {
	Message string
}

// greeterServiceName is the fully-qualified gRPC service name used in method
// routing.
const greeterServiceName = "sample.Greeter"

// Server is a minimal gRPC server wrapping grpc.Server with graceful.Service
// semantics (Start blocks until Shutdown closes the listener).
type Server struct {
	log    *zap.SugaredLogger
	addr   string
	server *grpc.Server
	ln     net.Listener
}

// New creates a gRPC server bound to addr with the Greeter service registered.
// Call Start to serve; Shutdown stops the server gracefully.
func New(log *zap.SugaredLogger, addr string, greeter GreeterServer) (*Server, error) {
	if log == nil {
		return nil, fmt.Errorf("grpc: nil logger")
	}
	if addr == "" {
		return nil, fmt.Errorf("grpc: empty listen address")
	}
	if greeter == nil {
		return nil, fmt.Errorf("grpc: nil greeter")
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("grpc: listen %s: %w", addr, err)
	}

	srv := grpc.NewServer()
	registerGreeter(srv, greeter)

	return &Server{
		log:    log,
		addr:   ln.Addr().String(),
		server: srv,
		ln:     ln,
	}, nil
}

// Addr returns the resolved listen address (useful when binding to :0).
func (s *Server) Addr() string { return s.addr }

// Start serves gRPC requests on the bound listener. It blocks until Shutdown
// is invoked, implementing graceful.Service.
func (s *Server) Start(_ context.Context) error {
	s.log.Infof("grpc server listening on %s", s.addr)
	if err := s.server.Serve(s.ln); err != nil {
		// GracefulStop / Stop cause Serve to return nil or a closed-listener
		// error; treat the latter as expected during shutdown.
		return fmt.Errorf("grpc: serve: %w", err)
	}
	return nil
}

// Shutdown stops the gRPC server. It implements graceful.Service.
//
// Uses Stop (immediate) rather than GracefulStop so shutdown is deterministic
// even when in-flight RPCs are stuck; in-flight RPCs receive a Canceled status.
// For workloads that must drain, swap in GracefulStop with a deadline.
func (s *Server) Shutdown(_ context.Context) error {
	s.log.Info("grpc server stopping")
	s.server.Stop()
	return nil
}
