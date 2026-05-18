package server

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

func TestRegisterServeGracefulStopRoundTrip(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	srv := New()
	srv.RegisterService(&testServiceDesc, testService{})

	errs := make(chan error, 1)
	go func() {
		errs <- srv.Serve(lis)
	}()

	conn := dialBufConn(t, lis)
	defer conn.Close()

	if err := conn.Invoke(t.Context(), "/plumego.rpc.test.Test/Ping", &emptypb.Empty{}, &emptypb.Empty{}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	stopCtx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := srv.GracefulStop(stopCtx); err != nil {
		t.Fatalf("GracefulStop() error = %v", err)
	}

	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve() did not exit after GracefulStop")
	}
}

func TestGracefulStopCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := New().GracefulStop(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("GracefulStop() error = %v, want %v", err, context.Canceled)
	}
}

func TestServeClosedListenerReturnsError(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	if err := lis.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if err := New().Serve(lis); err == nil {
		t.Fatal("Serve() error = nil, want error")
	}
}

func TestUnaryInterceptorIsInvoked(t *testing.T) {
	var invoked atomic.Bool
	unary := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		invoked.Store(true)
		return handler(ctx, req)
	}

	lis := bufconn.Listen(bufSize)
	srv := New(WithInterceptors(unary, nil)...)
	srv.RegisterService(&testServiceDesc, testService{})

	errs := make(chan error, 1)
	go func() {
		errs <- srv.Serve(lis)
	}()
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.GracefulStop(stopCtx); err != nil {
			t.Fatalf("GracefulStop() error = %v", err)
		}
		if err := <-errs; err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	}()

	conn := dialBufConn(t, lis)
	defer conn.Close()

	if err := conn.Invoke(t.Context(), "/plumego.rpc.test.Test/Ping", &emptypb.Empty{}, &emptypb.Empty{}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if !invoked.Load() {
		t.Fatal("expected unary interceptor to be invoked")
	}
}

func dialBufConn(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "passthrough:///bufconn",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}
	return conn
}

type pingService interface {
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

type testService struct{}

func (testService) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

var testServiceDesc = grpc.ServiceDesc{
	ServiceName: "plumego.rpc.test.Test",
	HandlerType: (*pingService)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    pingHandler,
		},
	},
}

func pingHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(pingService).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plumego.rpc.test.Test/Ping",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		event, ok := req.(*emptypb.Empty)
		if !ok {
			return nil, status.Error(codes.Internal, "unexpected request type")
		}
		return srv.(pingService).Ping(ctx, event)
	}
	return interceptor(ctx, in, info, handler)
}
