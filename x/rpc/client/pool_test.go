package client

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	plumelog "github.com/spcent/plumego/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

func TestPoolDialCachesConnection(t *testing.T) {
	lis, cleanup := startTestServer(t, testService{})
	defer cleanup()

	pool := New(bufDialOptions(lis)...)
	defer pool.Close()

	first, err := pool.Dial(t.Context(), "passthrough:///bufconn")
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	second, err := pool.Dial(t.Context(), "passthrough:///bufconn")
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if first != second {
		t.Fatal("Dial() returned different connections for the same target")
	}
}

func TestPoolDialAfterCloseReturnsError(t *testing.T) {
	pool := New()
	if err := pool.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := pool.Dial(t.Context(), "passthrough:///closed"); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("Dial() error = %v, want %v", err, ErrPoolClosed)
	}
}

func TestLoggingInterceptorLogsMethod(t *testing.T) {
	lis, cleanup := startTestServer(t, testService{})
	defer cleanup()

	logger := &recordingLogger{}
	pool := New(append(bufDialOptions(lis), grpc.WithUnaryInterceptor(LoggingInterceptor(logger)))...)
	defer pool.Close()

	conn, err := pool.Dial(t.Context(), "passthrough:///bufconn")
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if err := conn.Invoke(t.Context(), testMethod, &emptypb.Empty{}, &emptypb.Empty{}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if logger.method.Load() != testMethod {
		t.Fatalf("logged method = %q, want %q", logger.method.Load(), testMethod)
	}
}

func TestRetryInterceptorRetriesUnavailable(t *testing.T) {
	var calls atomic.Int64
	lis, cleanup := startTestServer(t, testService{
		ping: func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
			if calls.Add(1) < 3 {
				return nil, status.Error(codes.Unavailable, "temporary")
			}
			return &emptypb.Empty{}, nil
		},
	})
	defer cleanup()

	pool := New(append(bufDialOptions(lis), grpc.WithUnaryInterceptor(RetryInterceptor(3, codes.Unavailable)))...)
	defer pool.Close()

	conn, err := pool.Dial(t.Context(), "passthrough:///bufconn")
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if err := conn.Invoke(t.Context(), testMethod, &emptypb.Empty{}, &emptypb.Empty{}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func bufDialOptions(lis *bufconn.Listener) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
}

func startTestServer(t *testing.T, svc testService) (*bufconn.Listener, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	srv.RegisterService(&testServiceDesc, svc)
	errs := make(chan error, 1)
	go func() {
		errs <- srv.Serve(lis)
	}()
	return lis, func() {
		srv.GracefulStop()
		select {
		case err := <-errs:
			if err != nil {
				t.Fatalf("Serve() error = %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("Serve() did not stop")
		}
	}
}

const testMethod = "/plumego.rpc.client.Test/Ping"

type pingService interface {
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

type testService struct {
	ping func(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

func (s testService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if s.ping != nil {
		return s.ping(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

var testServiceDesc = grpc.ServiceDesc{
	ServiceName: "plumego.rpc.client.Test",
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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: testMethod}
	handler := func(ctx context.Context, req any) (any, error) {
		empty, ok := req.(*emptypb.Empty)
		if !ok {
			return nil, status.Error(codes.Internal, "unexpected request type")
		}
		return srv.(pingService).Ping(ctx, empty)
	}
	return interceptor(ctx, in, info, handler)
}

type recordingLogger struct {
	method atomic.Value
}

func (l *recordingLogger) WithFields(plumelog.Fields) plumelog.StructuredLogger { return l }
func (l *recordingLogger) Debug(string, ...plumelog.Fields)                     {}
func (l *recordingLogger) Warn(string, ...plumelog.Fields)                      {}
func (l *recordingLogger) Error(string, ...plumelog.Fields)                     {}
func (l *recordingLogger) Fatal(string, ...plumelog.Fields)                     {}
func (l *recordingLogger) DebugCtx(context.Context, string, ...plumelog.Fields) {}
func (l *recordingLogger) WarnCtx(context.Context, string, ...plumelog.Fields)  {}
func (l *recordingLogger) ErrorCtx(context.Context, string, ...plumelog.Fields) {}
func (l *recordingLogger) FatalCtx(context.Context, string, ...plumelog.Fields) {}

func (l *recordingLogger) Info(_ string, fields ...plumelog.Fields) {
	l.storeMethod(fields...)
}

func (l *recordingLogger) InfoCtx(_ context.Context, _ string, fields ...plumelog.Fields) {
	l.storeMethod(fields...)
}

func (l *recordingLogger) storeMethod(fields ...plumelog.Fields) {
	for _, fieldSet := range fields {
		if method, ok := fieldSet["method"].(string); ok {
			l.method.Store(method)
			return
		}
	}
}
