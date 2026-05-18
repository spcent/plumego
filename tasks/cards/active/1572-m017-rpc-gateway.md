# Card 1572

Milestone: M-017
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: active
Primary Module: x/rpc
Owned Files:
- `x/rpc/gateway/transcoder.go`
- `x/rpc/gateway/transcoder_test.go`
- `reference/with-rpc/main.go`
- `reference/with-rpc/go.mod`

Goal:
- Create x/rpc/gateway/ with an HTTPTranscoder that registers gRPC service
  descriptors as HTTP routes in a plumego router, and ship reference/with-rpc
  as a minimal end-to-end example.

Scope:
- Create x/rpc/gateway/transcoder.go defining:
  - HTTPTranscoder struct holding a grpc-gateway runtime.ServeMux and target
    gRPC address.
  - New(target string, opts ...runtime.ServeMuxOption) *HTTPTranscoder.
  - Register(ctx context.Context, handler runtime.HandlerFunc, pattern string)
    — registers a transcoded endpoint.
  - Handler() http.Handler — returns the ServeMux as an http.Handler for use
    with router.AddRoute or core.App.
- Create x/rpc/gateway/transcoder_test.go using httptest and bufconn:
  - Register + Handler returns 200 for a valid transcoded request.
  - Unknown path returns 404.
  - gRPC error maps to correct HTTP status code.
- Create reference/with-rpc/go.mod and reference/with-rpc/main.go:
  - Defines a minimal HelloService gRPC service (no protobuf required; use
    manual descriptor or a checked-in test proto).
  - Wires x/rpc/server.Server and x/rpc/gateway.HTTPTranscoder with core.App.
  - Demonstrates graceful shutdown: both gRPC and HTTP servers stop on context
    cancellation.

Non-goals:
- Do not implement protobuf code generation; use a manually defined descriptor.
- Do not add grpc-gateway to the main module.
- Do not require a real network port for tests.

Files:
- `x/rpc/gateway/transcoder.go`
- `x/rpc/gateway/transcoder_test.go`
- `reference/with-rpc/main.go`
- `reference/with-rpc/go.mod`

Tests:
- `go test -race -timeout 60s ./x/rpc/gateway/...`
- `go vet ./x/rpc/...`
- `go build ./reference/with-rpc/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Update x/rpc/module.yaml doc_paths to include reference/with-rpc.
- Add reference/with-rpc to the reference/ README.

Done Definition:
- HTTPTranscoder.Handler() integrates with router.AddRoute.
- All three transcoder_test.go cases pass.
- reference/with-rpc builds and runs without a real network port.
- `go run ./internal/checks/reference-layout` exits 0.

Outcome:
-
