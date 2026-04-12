# Card 0949: Middleware Observability Request Lifecycle Canonicalization

Priority: P1
State: active
Primary Module: middleware

## Goal

Establish one canonical request-lifecycle path for stable observability middleware so request ID preparation, response recording, span correlation, and HTTP metrics collection do not drift across `accesslog`, `tracing`, and `httpmetrics`.

## Problem

- `middleware/accesslog/accesslog.go`, `middleware/tracing/tracing.go`, and `middleware/httpmetrics/http_metrics.go` all start from `middleware/internal/observability.PrepareRequest(...)` and `BuildRequestMetrics(...)`.
- Even with the shared helper, each middleware still owns a slightly different slice of the lifecycle:
  - request ID stamping
  - response recorder setup
  - span extraction / `X-Span-ID` attachment
  - HTTP metric path selection
  - end-of-request cleanup
- This creates two concrete consistency risks:
  - request ID generation happens implicitly inside observability middleware even when `middleware/requestid` is not in the chain, so ownership is no longer obvious from route wiring
  - access logging and tracing each carry their own span-finalization glue, which makes ordering and double-instrumentation behavior harder to reason about
- The current tests pass, but the surface is mechanically duplicated enough that future edits are likely to diverge.

## Scope

- Refactor stable observability middleware so request preparation and completion flow through one explicit lifecycle contract.
- Make request ID ownership semantics explicit:
  - either observability middleware intentionally owns fallback generation everywhere
  - or `requestid.Middleware()` becomes the only generator and observability middleware becomes read-only
- Consolidate shared completion logic for span correlation and HTTP metrics path selection where it is currently copied across accesslog and tracing.
- Add focused regression tests for the chosen request ID policy and for tracer / metrics / access-log interaction order.

## Non-Goals

- Do not introduce `x/observability` dependencies into stable middleware.
- Do not change access-log field names or metrics wire contracts unless the card explicitly requires a documented behavior change.
- Do not redesign the public `Tracer` or `HTTPObserver` interfaces.

## Files

- `middleware/internal/observability/helpers.go`
- `middleware/accesslog/accesslog.go`
- `middleware/tracing/tracing.go`
- `middleware/httpmetrics/http_metrics.go`
- `middleware/*_test.go`

## Tests

- `go test -timeout 20s ./middleware/...`
- `go test -race -timeout 60s ./middleware/...`
- `go vet ./middleware/...`

## Docs Sync

- Sync `docs/modules/middleware/README.md` if the canonical request ID ownership story becomes stricter or more explicit.

## Done Definition

- Stable observability middleware shares one explicit request lifecycle instead of parallel partial implementations.
- Request ID generation ownership is documented by the code path and covered by tests.
- Access log, tracing, and HTTP metrics middleware no longer duplicate span/request completion logic unnecessarily.
- Middleware tests cover the chosen ordering and fallback behavior.
