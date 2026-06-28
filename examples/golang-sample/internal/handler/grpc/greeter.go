package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Greeter is the default GreeterServer implementation. It builds a greeting
// from the incoming name.
type Greeter struct {
	log *zap.SugaredLogger
}

// Compile-time guard.
var _ GreeterServer = (*Greeter)(nil)

// NewGreeter constructs a default Greeter.
func NewGreeter(log *zap.SugaredLogger) *Greeter {
	return &Greeter{log: log}
}

// SayHello returns a greeting for the named caller.
func (g *Greeter) SayHello(_ context.Context, req *HelloRequest) (*HelloReply, error) {
	if req == nil {
		return nil, fmt.Errorf("greeter: nil request")
	}
	name := req.Name
	if name == "" {
		name = "stranger"
	}
	g.log.Debugf("SayHello from %q", name)
	return &HelloReply{Message: "Hello, " + name}, nil
}
