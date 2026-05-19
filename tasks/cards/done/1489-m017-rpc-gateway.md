# Card 1572

Milestone: M-017
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: done
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
- Added `x/rpc/gateway.HTTPTranscoder` over grpc-gateway runtime ServeMux with
  explicit target storage, HTTP pattern registration, and `http.Handler`
  integration for core.App/router mounting.
- Added gateway tests with bufconn-backed gRPC calls covering a registered 200
  route, unknown-path 404, and gRPC status to HTTP status mapping.
- Added `reference/with-rpc`, a no-real-port example that wires
  `x/rpc/server.Server`, `x/rpc/gateway.HTTPTranscoder`, and `core.App`, runs a
  sample request through httptest, and shuts down HTTP and gRPC lifecycles via
  context.
- Updated x/rpc module metadata, dependency rules, taxonomy/hotspot docs, and
  docs index entries for the new gateway and reference example.
- Validation:
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go test -race
  -timeout 60s ./gateway/...` from x/rpc; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go vet ./...` from x/rpc;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go test -timeout
  20s ./...` from x/rpc; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go build ./...` from x/rpc;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go build ./...`
  from reference/with-rpc; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./...` from reference/with-rpc;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/reference-layout`; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/dependency-rules`;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/module-manifests`; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/agent-workflow`;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/public-entrypoints-sync`; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-maturity`;
  `gofmt -l .`; `git diff --check`.
