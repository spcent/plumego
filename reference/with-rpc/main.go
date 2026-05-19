package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spcent/plumego/core"
	rpcgateway "github.com/spcent/plumego/x/rpc/gateway"
	rpcserver "github.com/spcent/plumego/x/rpc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	bufSize     = 1024 * 1024
	helloMethod = "/plumego.reference.rpc.HelloService/Hello"
)

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis := bufconn.Listen(bufSize)
	grpcServer := rpcserver.New(grpcRuntime{server: grpc.NewServer()})
	if err := grpcServer.RegisterService(&helloServiceDesc, helloService{}); err != nil {
		return fmt.Errorf("register grpc service: %w", err)
	}
	grpcDone := make(chan error, 1)
	go func() {
		grpcDone <- grpcServer.Serve(lis)
	}()

	conn, err := grpc.DialContext(ctx, "passthrough:///bufconn",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("dial grpc server: %w", err)
	}

	transcoder := rpcgateway.New("bufconn")
	if err := transcoder.Register(ctx, adaptRuntimeHandler(helloGatewayHandler(conn)), "GET /v1/hello"); err != nil {
		return fmt.Errorf("register transcoder: %w", err)
	}

	cfg := core.DefaultConfig()
	cfg.Addr = "127.0.0.1:0"
	app := core.New(cfg, core.AppDependencies{})
	if err := app.Get("/v1/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transcoder.Handler().ServeHTTP(w, r)
	})); err != nil {
		return fmt.Errorf("register http route: %w", err)
	}
	if err := app.Prepare(); err != nil {
		return fmt.Errorf("prepare http app: %w", err)
	}
	srv, err := app.Server()
	if err != nil {
		return fmt.Errorf("prepare http server: %w", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return fmt.Errorf("unexpected http status %d: %s", rec.Code, rec.Body.String())
	}
	fmt.Print(rec.Body.String())

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := app.Shutdown(stopCtx); err != nil {
		return fmt.Errorf("shutdown http app: %w", err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("close grpc client: %w", err)
	}
	if err := grpcServer.GracefulStop(stopCtx); err != nil {
		return fmt.Errorf("shutdown grpc server: %w", err)
	}
	if err := <-grpcDone; err != nil {
		return fmt.Errorf("grpc serve: %w", err)
	}
	return nil
}

func helloGatewayHandler(conn *grpc.ClientConn) runtime.HandlerFunc {
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

func adaptRuntimeHandler(handler runtime.HandlerFunc) rpcgateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		handler(w, r, params)
	}
}

type grpcRuntime struct {
	server *grpc.Server
}

func (r grpcRuntime) RegisterService(desc any, impl any) error {
	serviceDesc, ok := desc.(*grpc.ServiceDesc)
	if !ok {
		return fmt.Errorf("unexpected service descriptor %T", desc)
	}
	r.server.RegisterService(serviceDesc, impl)
	return nil
}

func (r grpcRuntime) Serve(lis net.Listener) error {
	return r.server.Serve(lis)
}

func (r grpcRuntime) GracefulStop() {
	r.server.GracefulStop()
}

func (r grpcRuntime) Stop() {
	r.server.Stop()
}

type helloServiceAPI interface {
	Hello(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
}

type helloService struct{}

func (helloService) Hello(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String("hello from rpc"), nil
}

var helloServiceDesc = grpc.ServiceDesc{
	ServiceName: "plumego.reference.rpc.HelloService",
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
