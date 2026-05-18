package gateway

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const bufSize = 1024 * 1024

func TestRegisterHandlerReturnsOK(t *testing.T) {
	conn, cleanup := startGatewayTestServer(t, helloService{})
	defer cleanup()

	transcoder := New("bufconn")
	err := transcoder.Register(t.Context(), helloHandler(conn), "GET /v1/hello")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	transcoder.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["message"] != "hello" {
		t.Fatalf("message = %q, want hello", body["message"])
	}
}

func TestUnknownPathReturnsNotFound(t *testing.T) {
	transcoder := New("bufconn")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)

	transcoder.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGRPCErrorMapsToHTTPStatus(t *testing.T) {
	conn, cleanup := startGatewayTestServer(t, helloService{
		hello: func(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
			return nil, status.Error(codes.PermissionDenied, "denied")
		},
	})
	defer cleanup()

	transcoder := New("bufconn")
	err := transcoder.Register(t.Context(), helloHandler(conn), "GET /v1/hello")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	transcoder.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func helloHandler(conn *grpc.ClientConn) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		var out wrapperspb.StringValue
		err := conn.Invoke(r.Context(), helloMethod, &emptypb.Empty{}, &out)
		if err != nil {
			http.Error(w, status.Convert(err).Message(), runtime.HTTPStatusFromCode(status.Code(err)))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": out.Value})
	}
}

func startGatewayTestServer(t *testing.T, svc helloService) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	srv.RegisterService(&helloServiceDesc, svc)
	errs := make(chan error, 1)
	go func() {
		errs <- srv.Serve(lis)
	}()

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

	return conn, func() {
		_ = conn.Close()
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

const helloMethod = "/plumego.rpc.gateway.HelloService/Hello"

type helloServiceAPI interface {
	Hello(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

type helloService struct {
	hello func(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

func (s helloService) Hello(ctx context.Context, req *emptypb.Empty) (*wrapperspb.StringValue, error) {
	if s.hello != nil {
		return s.hello(ctx, req)
	}
	return wrapperspb.String("hello"), nil
}

var helloServiceDesc = grpc.ServiceDesc{
	ServiceName: "plumego.rpc.gateway.HelloService",
	HandlerType: (*helloServiceAPI)(nil),
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
		return srv.(helloServiceAPI).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: helloMethod}
	handler := func(ctx context.Context, req any) (any, error) {
		empty, ok := req.(*emptypb.Empty)
		if !ok {
			return nil, status.Error(codes.Internal, "unexpected request type")
		}
		return srv.(helloServiceAPI).Hello(ctx, empty)
	}
	return interceptor(ctx, in, info, handler)
}
