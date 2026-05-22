// Package grpc contains the gRPC service definition, implementation, and
// the runtime adapter that bridges grpc.Server to the rpcserver interface.
package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	ServiceName = "plumego.reference.rpc.HelloService"
	HelloMethod = "/" + ServiceName + "/Hello"
)

// HelloServiceAPI is the server-side gRPC interface for the HelloService.
type HelloServiceAPI interface {
	Hello(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

// HelloService is the canonical offline implementation of HelloServiceAPI.
type HelloService struct{}

// Hello returns a static greeting to demonstrate gRPC → HTTP transcoding.
func (HelloService) Hello(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String("hello from rpc"), nil
}

// HelloServiceDesc is the gRPC service descriptor for HelloService.
// It registers the Hello unary method without generated protobuf code.
var HelloServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*HelloServiceAPI)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    helloUnaryHandler,
		},
	},
}

func helloUnaryHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServiceAPI).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: HelloMethod}
	handler := func(ctx context.Context, req any) (any, error) {
		empty, ok := req.(*emptypb.Empty)
		if !ok {
			return nil, status.Error(codes.Internal, "unexpected request type")
		}
		return srv.(HelloServiceAPI).Hello(ctx, empty)
	}
	return interceptor(ctx, in, info, handler)
}

// Runtime adapts *grpc.Server to the rpcserver.Runtime interface.
type Runtime struct {
	Server *grpc.Server
}

// RegisterService registers a service descriptor and implementation on the
// underlying grpc.Server.
func (r Runtime) RegisterService(desc any, impl any) error {
	serviceDesc, ok := desc.(*grpc.ServiceDesc)
	if !ok {
		return fmt.Errorf("unexpected service descriptor %T", desc)
	}
	r.Server.RegisterService(serviceDesc, impl)
	return nil
}

// Serve accepts incoming connections on the listener.
func (r Runtime) Serve(lis net.Listener) error {
	return r.Server.Serve(lis)
}

// GracefulStop shuts down the gRPC server gracefully.
func (r Runtime) GracefulStop() {
	r.Server.GracefulStop()
}

// Stop terminates the gRPC server immediately.
func (r Runtime) Stop() {
	r.Server.Stop()
}
