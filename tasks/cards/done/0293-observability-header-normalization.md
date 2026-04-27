# Card 0293: Observability Header Normalization

Priority: P2
State: done
Primary Module: middleware

## Goal

Make observability header behavior uniform across access-log, tracing, and shared request preparation helpers. Eliminate repeated string literals and align header injection on a single helper path.

## Problem

`middleware/accesslog` and `middleware/tracing` both hardcode `"X-Span-ID"` and inject span headers independently. `middleware/internal/observability.PrepareRequest` sets request-id context and response headers directly, while `AttachRequestID` handles the same concern elsewhere. This results in duplicated behavior and makes it easy for future middleware to drift in how span/request headers are applied.

## Scope

- Introduce a single canonical constant for the span header name used by middleware.
- Route span header injection through a shared helper instead of duplicating `w.Header().Set("X-Span-ID", ...)`.
- Align `PrepareRequest` to use the existing request-id attachment helper so request-id behavior is consistent.
- Update access-log and tracing middleware to use the shared helper/constant.
- Update tests to assert on the canonical header constant.

## Non-Goals

- Do not change trace context propagation semantics or introduce new trace headers.
- Do not change request-id format or generation policy.
- Do not move tracing infrastructure out of `middleware`.

## Expected Files

- `middleware/internal/observability/*.go`
- `middleware/accesslog/accesslog.go`
- `middleware/tracing/tracing.go`
- `middleware/*/*_test.go`

## Validation

```bash
go test -timeout 20s ./middleware/...
go test -race -timeout 60s ./middleware/...
go vet ./middleware/...
```

## Done Definition

- Span header name is defined once and referenced from all middleware that sets it.
- Request-id attachment is done through one helper path.
- No direct `"X-Span-ID"` literals remain in middleware packages.

## Outcome

- Added `SpanIDHeader` and `AttachSpanID` in internal observability helpers and routed tracing/access-log through them.
- `PrepareRequest` now uses `AttachRequestID` to unify request-id attachment behavior.
- Updated span header assertions in tracing tests.
- Validation: `go test -timeout 20s ./middleware/...`, `go test -race -timeout 60s ./middleware/...`, `go vet ./middleware/...`.
