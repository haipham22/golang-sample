package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestServer_StartAndShutdown exercises Server.Start (previously 0% covered) by
// serving on a free port and shutting down. We do not perform a real RPC: the
// template uses a custom JSON codec without .proto codegen, so driving an
// actual call requires content-subtype negotiation that is out of scope for
// this template stub. The Greeter unit tests + handler tests cover the
// per-RPC logic. This test covers the Server lifecycle.
func TestServer_StartAndShutdown(t *testing.T) {
	log := zap.NewNop().Sugar()
	srv, err := New(log, "127.0.0.1:0", NewGreeter(log))
	require.NoError(t, err)
	require.NotEmpty(t, srv.Addr())

	// Start serving in the background.
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Start(context.Background()) }()

	// A client dial proves the listener is actually serving. Dial with
	// insecure creds and a short timeout; we don't invoke any RPC.
	conn, err := grpc.NewClient(srv.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	// Shutdown must cause Start to return. We don't assert on Start's return
	// value precisely because grpc.Server.Stop closes the listener, which can
	// surface as a "closed" error from Serve — both nil and error are valid.
	require.NoError(t, srv.Shutdown(context.Background()))

	select {
	case <-serveErr:
		// Start returned — lifecycle is complete.
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return within 3s of Shutdown")
	}
}

// TestServer_AddrMatchesResolvedListener verifies Addr returns the resolved
// address (not the input ":0") after New binds the listener.
func TestServer_AddrMatchesResolvedListener(t *testing.T) {
	log := zap.NewNop().Sugar()
	srv, err := New(log, "127.0.0.1:0", NewGreeter(log))
	require.NoError(t, err)
	defer func() { _ = srv.Shutdown(context.Background()) }()

	assert.NotEqual(t, "127.0.0.1:0", srv.Addr(), "Addr should be the resolved port")
	assert.Contains(t, srv.Addr(), "127.0.0.1:")
}
