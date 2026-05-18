# Plan for M-017: gRPC Support

Milestone: `M-017`
Objective: Ship x/rpc with server, client, and gateway sub-packages that
integrate gRPC lifecycle with core.App and connect to x/gateway for HTTP/JSON
transcoding, with offline tests and a hello-world reference app.
Constraints: x/rpc has its own go.mod with google.golang.org/grpc dependency,
no gRPC import in main module or stable roots, no running gRPC service required
for unit tests, protobuf code generation is the user's responsibility.
Affected Modules: x/rpc, x/gateway (read-only), reference/with-rpc.

## Phase Map

- Phase 1: Orient — read core/app.go for lifecycle contract, x/gateway/module.yaml
  for transcoding interface, and x/observability/module.yaml for tracing interface.
- Phase 2: Implement (parallel) — write server, client pool, and gateway transcoder
  concurrently since they are independent sub-packages.
- Phase 3: Test — write offline tests for all three sub-packages using in-process
  loopback and mock backends.
- Phase 4: Reference and Validate — add reference/with-rpc, run acceptance criteria,
  commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1570 | Create x/rpc/server/ wrapping grpc.Server lifecycle | x/rpc | `x/rpc/server/server.go`, `x/rpc/server/server_test.go`, `x/rpc/go.mod`, `x/rpc/module.yaml` | M-014 | `go test ./x/rpc/server/...`, `go vet ./x/rpc/...` |
| 1571 | Create x/rpc/client/ with Pool and interceptors | x/rpc | `x/rpc/client/client.go`, `x/rpc/client/client_test.go` | 1570 | `go test ./x/rpc/client/...` |
| 1572 | Create x/rpc/gateway/ with HTTPTranscoder | x/rpc | `x/rpc/gateway/gateway.go`, `x/rpc/gateway/gateway_test.go` | 1570 | `go test ./x/rpc/gateway/...` |

## Dependency Edges

- `1570 -> 1571`
- `1570 -> 1572`

## Parallel Groups

- Group A: card 1570 — must complete first; creates x/rpc/go.mod and the server package
  that establishes the module boundary.
- Group B (parallel after A): cards 1571 and 1572 — independent sub-packages sharing
  the go.mod created in 1570.
- Group C (sequential): reference/with-rpc and acceptance criteria after all three cards.

## Risk Register

- Risk: grpc-gateway dependency version conflicts with grpc core version.
  Mitigation: pin both in x/rpc/go.mod using versions known to be compatible; record
  versions in x/rpc/module.yaml under deps.
- Risk: offline test for gateway requires a real protobuf descriptor to register routes.
  Mitigation: use a hand-written descriptor struct in the test; do not depend on
  proto code generation in CI.

## Verification Strategy

- Card-level checks: each sub-package card runs `go test` and `go vet` immediately.
- Module isolation check: `go run ./internal/checks/dependency-rules` confirms no
  gRPC dependency in the main module go.mod.
- Reference check: `go build ./reference/with-rpc/...` and `go test ./reference/with-rpc/...`
  must both exit 0.

## Exit Condition

- all three sub-package cards completed with offline tests
- x/rpc/go.mod is self-contained and separate from main module
- reference/with-rpc builds and tests pass
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
