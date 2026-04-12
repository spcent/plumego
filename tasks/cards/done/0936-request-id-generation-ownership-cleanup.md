# Card 0936: Request ID Generation Ownership Cleanup

Priority: P1
State: done
Primary Module: middleware

## Goal

Make request-id generation explicit and remove hidden global generators from the stable middleware surface.

## Problem

`middleware/internal/observability` exposes `NewRequestID()` backed by a package-global generator (`globalRequestIDGenerator`). This is a hidden global state with implicit lifecycle, while the public request-id middleware already supports explicit generator injection. The split makes generation policy unclear and violates the “no hidden globals” guidance.

## Scope

- Enumerate all call sites of `internal/observability.NewRequestID`.
- Move request-id generation ownership to a single public surface (`middleware/requestid`) with explicit constructors.
- Replace `internal/observability.NewRequestID` usage with injected or explicitly constructed generators.
- Remove the global generator or make it private to the request-id middleware in a way that is explicit at wiring time.
- Update tests for request-id generation and observability helpers.

## Non-Goals

- Do not change the request-id format.
- Do not change header names or request-id propagation semantics.
- Do not add new dependency on `x/*` from stable roots.

## Expected Files

- `middleware/internal/observability/request_id_generation.go`
- `middleware/requestid/*`
- `middleware/internal/observability/*`
- `middleware/*/*_test.go`

## Validation

```bash
go test -timeout 20s ./middleware/...
go test -race -timeout 60s ./middleware/...
go vet ./middleware/...
```

## Done Definition

- No hidden global request-id generator remains in stable middleware packages.
- All request-id generation flows through explicit constructors or injected generators.
- Middleware tests cover the new explicit generator path.

## Outcome

- Moved request-id generator + decode helpers into `middleware/requestid` with explicit `DefaultGenerator`.
- Consolidated request-id header helpers in `middleware/requestid` and routed observability/request-id middleware through them.
- Validation: `go test -timeout 20s ./middleware/...`, `go test -race -timeout 60s ./middleware/...`, `go vet ./middleware/...`.
