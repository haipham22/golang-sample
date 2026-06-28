package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGreeter_SayHello(t *testing.T) {
	t.Run("named caller", func(t *testing.T) {
		g := NewGreeter(zap.NewNop().Sugar())
		resp, err := g.SayHello(context.Background(), &HelloRequest{Name: "Alice"})
		require.NoError(t, err)
		assert.Equal(t, "Hello, Alice", resp.Message)
	})

	t.Run("empty name defaults to stranger", func(t *testing.T) {
		g := NewGreeter(zap.NewNop().Sugar())
		resp, err := g.SayHello(context.Background(), &HelloRequest{Name: ""})
		require.NoError(t, err)
		assert.Equal(t, "Hello, stranger", resp.Message)
	})

	t.Run("nil request rejected", func(t *testing.T) {
		g := NewGreeter(zap.NewNop().Sugar())
		_, err := g.SayHello(context.Background(), nil)
		require.Error(t, err)
	})
}

func TestNew_Validation(t *testing.T) {
	t.Run("nil logger", func(t *testing.T) {
		_, err := New(nil, ":0", NewGreeter(zap.NewNop().Sugar()))
		require.Error(t, err)
	})
	t.Run("empty addr", func(t *testing.T) {
		_, err := New(zap.NewNop().Sugar(), "", NewGreeter(zap.NewNop().Sugar()))
		require.Error(t, err)
	})
	t.Run("nil greeter", func(t *testing.T) {
		_, err := New(zap.NewNop().Sugar(), ":0", nil)
		require.Error(t, err)
	})
}

// TestNew_BindsToFreePort verifies the server constructs and binds a listener
// on a free port. We do NOT exercise a live RPC here: the template uses a
// custom JSON codec without .proto codegen, and driving a real RPC requires
// matching content-subtype negotiation that is deliberately out of scope for
// this template stub. The Greeter service itself is covered by the unit tests
// above; a real project would add protoc-gen-go-grpc stubs and an end-to-end
// bufconn test.
func TestNew_BindsToFreePort(t *testing.T) {
	srv, err := New(zap.NewNop().Sugar(), "127.0.0.1:0", NewGreeter(zap.NewNop().Sugar()))
	require.NoError(t, err)
	require.NotEmpty(t, srv.Addr())
	// The listener is owned by the server; closing it via Shutdown is safe and
	// does not require Start to have been called.
	require.NoError(t, srv.Shutdown(context.Background()))
}

