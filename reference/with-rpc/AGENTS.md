# AGENTS.md - reference/with-rpc

Operational guide for agents working in `reference/with-rpc`.

This file is intentionally short. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

This service differs from other references: there is no `internal/config` package
and no runtime configuration. Read `internal/app/app.go` directly.

## 2. Purpose And Boundaries

`reference/with-rpc` demonstrates how to run a gRPC server alongside an HTTP/JSON
transcoding gateway in the same process using `x/rpc`. The gRPC server and HTTP
gateway communicate via an in-process `bufconn` connection — no network port is
allocated for gRPC. The HTTP gateway exposes `GET /v1/hello` as a JSON endpoint
that transcodes to a gRPC call on `HelloService.HelloMethod`.

Hard rules:

- `x/rpc/gateway` and `x/rpc/server` are the only allowed `x/*` imports.
- No new third-party dependencies beyond the existing grpc-gateway dependency.
- No hidden globals, `init()` registration, or reflection routing.
- Keep the gRPC service implementation in a dedicated package.
- The HTTP endpoint shape (`GET /v1/hello`) is part of the demo contract — do
  not change it without updating the test.
- This service has no runtime config; do not add one without a clear reason.

## 3. Package Ownership

- `main.go`: process entrypoint; constructs and starts the app.
- `internal/app/app.go`: gRPC server setup (bufconn), transcoder, HTTP server
  construction, graceful shutdown.
- `internal/hello`: gRPC service implementation (`HelloService.HelloMethod`).

Dependency direction:

```text
main.go → app → hello, x/rpc/server, x/rpc/gateway
```

## 4. Change Patterns

Add a gRPC method:

1. Add the method to the gRPC service interface and implementation in `internal/hello`.
2. Register the HTTP/JSON transcoding route in `app.go` using the gateway.
3. Add a focused test that dials the bufconn and asserts the JSON response.

Change the response shape:

1. Update `internal/hello` to return the new fields.
2. Update the test assertions in `internal/app/app_test.go`.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-rpc && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- gRPC server uses in-process bufconn — no external port is allocated for gRPC.
- `x/rpc/gateway` and `x/rpc/server` only — no other `x/*` packages.
- HTTP response shape matches the transcoded proto response.
- Graceful shutdown closes the bufconn listener and gRPC server before HTTP.
