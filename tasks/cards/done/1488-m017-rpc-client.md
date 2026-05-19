# Card 1571

Milestone: M-017
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: done
Primary Module: x/rpc
Owned Files:
- `x/rpc/client/pool.go`
- `x/rpc/client/pool_test.go`
- `x/rpc/client/interceptors.go`

Goal:
- Create x/rpc/client/ with a connection pool type and built-in unary
  interceptors for structured logging, distributed tracing, and linear retry.

Scope:
- Create x/rpc/client/pool.go defining:
  - Pool struct managing a map of target → *grpc.ClientConn.
  - New(defaultOpts ...grpc.DialOption) *Pool — constructor.
  - Dial(ctx context.Context, target string, opts ...grpc.DialOption)
    (*grpc.ClientConn, error) — returns cached conn or dials a new one.
  - Close() error — closes all pooled connections.
  - WithKeepAlive(params keepalive.ClientParameters) DialOption helper.
- Create x/rpc/client/interceptors.go defining:
  - LoggingInterceptor(logger log.StructuredLogger) grpc.UnaryClientInterceptor
    — logs method, duration, and status code.
  - RetryInterceptor(maxAttempts int, codes ...codes.Code)
    grpc.UnaryClientInterceptor — linear retry on transient codes.
  - TracingInterceptor(tracer observability.Tracer) grpc.UnaryClientInterceptor
    — propagates trace context in outgoing metadata.
- Write x/rpc/client/pool_test.go using bufconn for in-process connection:
  - Dial same target twice returns same connection.
  - Dial after Close returns error.
  - LoggingInterceptor logs method name on call.
  - RetryInterceptor retries on codes.Unavailable up to maxAttempts.

Non-goals:
- Do not implement load balancing (use grpc's built-in round-robin resolver).
- Do not implement streaming interceptors in this card.
- Do not add circuit breaker to the client pool (that is x/resilience concern).

Files:
- `x/rpc/client/pool.go`
- `x/rpc/client/pool_test.go`
- `x/rpc/client/interceptors.go`

Tests:
- `go test -race -timeout 60s ./x/rpc/client/...`
- `go vet ./x/rpc/client/...`

Docs Sync:
- none at this card.

Done Definition:
- Pool.Dial returns cached connections.
- RetryInterceptor retries on transient codes up to maxAttempts.
- All four pool_test.go cases pass with `go test -race`.

Outcome:
- Added `x/rpc/client.Pool` with target-keyed `grpc.ClientConn` caching,
  closed-pool protection, aggregate close errors, and a keepalive dial option
  helper.
- Added unary client interceptors for structured logging, caller-selected retry
  on gRPC status codes, and trace context propagation through outgoing metadata.
- Added a `x/rpc/client` module manifest and updated `x/rpc`/dependency
  control-plane metadata for the new client subpackage.
- Added bufconn tests covering cached dials, dial-after-close errors, logging
  method fields, and retrying `codes.Unavailable` until success.
- Validation:
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go test -race
  -timeout 60s ./client/...` from x/rpc; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go vet ./client/...` from x/rpc;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go test -timeout
  20s ./...` from x/rpc; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go build ./...` from x/rpc;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/dependency-rules`; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/module-manifests`;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/agent-workflow`; `gofmt -l .`; `git diff --check`.
