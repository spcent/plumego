# Card 1570

Milestone: M-017
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: active
Primary Module: x/rpc
Owned Files:
- `x/rpc/server/server.go`
- `x/rpc/server/server_test.go`
- `x/rpc/server/module.yaml`
- `x/rpc/go.mod`

Goal:
- Create x/rpc/server/ wrapping grpc.Server with a lifecycle that aligns with
  core.App.Shutdown via context cancellation.

Scope:
- Create x/rpc/go.mod with module github.com/spcent/plumego/x/rpc and
  dependency google.golang.org/grpc.
- Create x/rpc/server/server.go defining:
  - Server struct wrapping *grpc.Server.
  - New(opts ...grpc.ServerOption) *Server — creates the underlying grpc.Server.
  - WithInterceptors(unary grpc.UnaryServerInterceptor,
    stream grpc.StreamServerInterceptor) Option — chains interceptors.
  - Serve(lis net.Listener) error — starts serving; blocks until stopped.
  - GracefulStop(ctx context.Context) error — calls grpc.Server.GracefulStop;
    returns ctx.Err() if context expires before graceful shutdown completes.
  - RegisterService(desc *grpc.ServiceDesc, impl any) — delegates to
    grpc.Server.RegisterService; must be called before Serve.
- Create x/rpc/server/module.yaml with status = experimental, forbidden_imports
  listing stable roots not on the allowed list and all x/* except observability.
- Write x/rpc/server/server_test.go using grpc/test/bufconn for in-process
  connection covering:
  - RegisterService + Serve + GracefulStop round-trip exits cleanly.
  - GracefulStop with already-cancelled context returns ctx.Err().
  - Serve on a closed listener returns error immediately.
  - Unary interceptor is invoked on request.

Non-goals:
- Do not import grpc in the main module or stable roots.
- Do not implement automatic service discovery registration.
- Do not require a real network listener for tests (use bufconn).

Files:
- `x/rpc/server/server.go`
- `x/rpc/server/server_test.go`
- `x/rpc/server/module.yaml`
- `x/rpc/go.mod`

Tests:
- `go test -race -timeout 60s ./x/rpc/server/...`
- `go vet ./x/rpc/server/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none at this card; reference app and primer added in card 1572.

Done Definition:
- x/rpc/go.mod is a separate module with google.golang.org/grpc.
- GracefulStop correctly handles context cancellation.
- All four server_test.go cases pass with `go test -race`.
- `go run ./internal/checks/dependency-rules` exits 0.

Outcome:
-
