# Card 1521

Milestone: —
Recipe: specs/change-recipes/add-acceptance-tests.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/standard-service
Owned Files:
- `reference/standard-service/internal/app/app_test.go`

## Goal

Add tests that verify `App.Start` graceful-shutdown lifecycle — the core production
capability that `reference/standard-service` claims to demonstrate but has zero
test coverage for today.

## Scope

New test functions in `internal/app/app_test.go` only.
Three behaviors to cover:

1. **Startup + real listen**: `App.Start` binds to `:0` (random port), serves one
   request, confirms the server is reachable.
2. **Graceful shutdown**: canceling the passed-in `context.Context` triggers shutdown;
   `App.Start` returns `nil` (not an unexpected error) within the 15-second drain window.
3. **Shutdown error propagation**: a synthesized shutdown error is propagated as the
   return value of `App.Start` (exercises the `<-shutdownErr` drain in `app.go:138`).

Each test must start and stop a real `net.Listener` using `:0`. Use
`httptest.NewServer` or `net.Listen("tcp", ":0")` + goroutine; do not hardcode a port.

## Non-goals

- Do not change `app.go`, `routes.go`, `main.go`, or any non-test file.
- Do not add TLS startup tests (separate card if needed; requires cert fixtures).
- Do not add tests for middleware behavior — that is card 1522.
- Do not change the module's `go.mod` or add new test dependencies.

## Files

- `reference/standard-service/internal/app/app_test.go`

## Acceptance Tests

```
reference/standard-service/internal/app/app_test.go: TestAcceptanceAppStartServesRequests
reference/standard-service/internal/app/app_test.go: TestAcceptanceAppStartGracefulShutdown
reference/standard-service/internal/app/app_test.go: TestAcceptanceAppStartPropagatesShutdownError
```

These functions must be written first and confirmed **failing** before implementation
begins (per `specs/change-recipes/add-acceptance-tests.yaml`).

## Tests

- Shutdown timeout: cancel ctx and confirm `App.Start` returns within 2 seconds
  for a server with no in-flight requests (no need to hit the 15-second wall).
- Confirm that a `200 OK` is received on the random port before cancellation.

## Docs Sync

None. This card adds tests only; it does not change documented behavior or public API.

## Validation

```
cd reference/standard-service && go test -race -timeout 60s ./internal/app/...
go run ./internal/checks/reference-layout
```

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l .` (inside `reference/standard-service`) produces no output.
- [ ] No new files created outside `internal/app/app_test.go`.

## Outcome

<!-- Agent fills this after completion. -->
