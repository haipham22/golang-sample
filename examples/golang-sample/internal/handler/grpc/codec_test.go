package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TestJSONCodec_RoundTrip verifies Marshal/Unmarshal round-trip a HelloRequest
// and that Name() reports the registered codec name. Covers the three
// jsonCodec methods which were previously 0%.
func TestJSONCodec_RoundTrip(t *testing.T) {
	c := jsonCodec{}

	assert.Equal(t, codecName, c.Name(), "Name() must match the registered name")

	original := &HelloRequest{Name: "alice"}
	data, err := c.Marshal(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	// JSON of {"Name":"alice"} — key is the Go field name.
	assert.Contains(t, string(data), `"Name":"alice"`)

	var decoded HelloRequest
	require.NoError(t, c.Unmarshal(data, &decoded))
	assert.Equal(t, "alice", decoded.Name)
}

// TestJSONCodec_UnmarshalRejectsBadJSON verifies the Unmarshal error path.
func TestJSONCodec_UnmarshalRejectsBadJSON(t *testing.T) {
	c := jsonCodec{}
	var req HelloRequest
	err := c.Unmarshal([]byte("{not-json"), &req)
	require.Error(t, err)
}

// TestJSONCodec_MarshalPropagatesError verifies Marshal surfaces JSON errors
// for non-encodable values (e.g. a channel).
func TestJSONCodec_MarshalPropagatesError(t *testing.T) {
	c := jsonCodec{}
	_, err := c.Marshal(make(chan int))
	require.Error(t, err)
}

// TestGreeterSayHelloHandler_HappyPath verifies the generated MethodHandler
// decodes the inbound payload via the dec func and dispatches to the
// GreeterServer. This is the path the gRPC runtime exercises on a real RPC.
func TestGreeterSayHelloHandler_HappyPath(t *testing.T) {
	impl := NewGreeter(zap.NewNop().Sugar())
	handler := greeterSayHelloHandler(impl)

	enc, _ := jsonCodec{}.Marshal(&HelloRequest{Name: "bob"})
	dec := func(v any) error {
		return jsonCodec{}.Unmarshal(enc, v)
	}

	resp, err := handler(impl, context.Background(), dec, nil)
	require.NoError(t, err)
	require.IsType(t, &HelloReply{}, resp)
	assert.Equal(t, "Hello, bob", resp.(*HelloReply).Message)
}

// TestGreeterSayHelloHandler_DecoderError verifies the handler returns the
// decoder error without invoking the Greeter (the 12.5%-covered branch).
func TestGreeterSayHelloHandler_DecoderError(t *testing.T) {
	impl := NewGreeter(zap.NewNop().Sugar())
	handler := greeterSayHelloHandler(impl)

	decErr := errors.New("decode boom")
	dec := func(any) error { return decErr }

	_, err := handler(impl, context.Background(), dec, nil)
	assert.ErrorIs(t, err, decErr)
}

// TestGreeterSayHelloHandler_ServerTypeMismatch verifies the handler rejects a
// server value that does not implement GreeterServer. Covers the remaining
// branch of greeterSayHelloHandler.
func TestGreeterSayHelloHandler_ServerTypeMismatch(t *testing.T) {
	handler := greeterSayHelloHandler(NewGreeter(zap.NewNop().Sugar()))

	dec := func(any) error { return nil } // empty payload still decodes to &HelloRequest{}
	_, err := handler("not-a-greeter", context.Background(), dec, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GreeterServer")
}

// TestRegisterGreeter_IsCallable ensures registerGreeter does not panic when
// handed a fresh grpc.Server. (It is already invoked by New, but exercising it
// directly here keeps it honest if New changes.)
func TestRegisterGreeter_IsCallable(t *testing.T) {
	srv := grpc.NewServer()
	registerGreeter(srv, NewGreeter(zap.NewNop().Sugar()))
	// No panic + server has a registered service. We assert via grpc.Server's
	// internal registry indirectly: GetServiceInfo should list our service.
	info := srv.GetServiceInfo()
	_, ok := info[greeterServiceName]
	assert.True(t, ok, "expected %q in service info", greeterServiceName)
}
