package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

// codecName identifies the custom JSON codec used by the sample Greeter
// service. Because the template deliberately avoids .proto codegen, requests
// and responses are framed as JSON on the wire via this codec. A real project
// generated from .proto would use the protobuf codec instead.
const codecName = "json"

// jsonCodec implements encoding.Codec for plain Go structs via encoding/json.
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)            { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error       { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                             { return codecName }

func init() {
	// Register the JSON codec so gRPC can (de)serialize our plain structs.
	// encoding.RegisterCodec replaces any prior codec with the same name.
	encoding.RegisterCodec(jsonCodec{})
}

// registerGreeter registers the sample Greeter service on srv using a manually
// constructed ServiceDesc (no protoc-generated code). The single SayHello
// method is exposed as a unary RPC under the json codec.
func registerGreeter(srv *grpc.Server, impl GreeterServer) {
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: greeterServiceName,
		HandlerType: (*GreeterServer)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "SayHello",
			Handler:    greeterSayHelloHandler(impl),
		}},
		Streams:  []grpc.StreamDesc{},
		Metadata: "greeter.proto",
	}, impl)
}

// greeterSayHelloHandler returns a grpc.MethodHandler that decodes the JSON
// request into *HelloRequest, calls the Greeter, and encodes the reply.
func greeterSayHelloHandler(_ GreeterServer) grpc.MethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		req := &HelloRequest{}
		if err := dec(req); err != nil {
			return nil, err
		}
		greeter, ok := srv.(GreeterServer)
		if !ok {
			return nil, fmt.Errorf("grpc: expected GreeterServer, got %T", srv)
		}
		return greeter.SayHello(ctx, req)
	}
}
