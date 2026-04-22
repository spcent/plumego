# Card 2104: Normalize x/ops Health HTTP Response Contract

Priority: P1
State: active
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Primary Module: x/ops
Owned Files:
- x/ops/healthhttp/handlers.go
- x/ops/healthhttp/readiness.go
- x/ops/healthhttp/history.go
- x/ops/healthhttp/runtime.go
- x/ops/healthhttp/metrics.go
Depends On:

## Goal

`x/ops/healthhttp` is an auth-gated operational HTTP surface, but it currently
has a separate success-response style from the rest of the repository:

- `handlers.go`, `readiness.go`, `history.go`, `runtime.go`, and `metrics.go`
  call `contract.WriteJSON` directly for JSON bodies.
- Some routes are bodyless health/status probes and can keep direct status
  writes, but JSON routes should have one documented shape.
- Error helpers already use `contract.WriteError`, so success and error paths
  are asymmetric.

## Scope

- Decide the canonical shape for `x/ops/healthhttp` JSON success bodies and
  implement it consistently.
- Prefer `contract.WriteResponse` for JSON bodies unless a health probe is
  intentionally raw/bodyless and tests document that exception.
- Preserve health status to HTTP status mapping.
- Keep protected ops behavior in `x/ops`; do not move HTTP exposure into the
  stable `health` root.
- Add tests that pin the response shape for health, readiness, history, runtime,
  and metrics routes.

## Non-goals

- Do not change `health.HealthStatus` domain types.
- Do not change auth policy or route registration outside `x/ops`.
- Do not alter runtime metric collection semantics.
- Do not add dependencies.

## Files

- `x/ops/healthhttp/handlers.go`
- `x/ops/healthhttp/readiness.go`
- `x/ops/healthhttp/history.go`
- `x/ops/healthhttp/runtime.go`
- `x/ops/healthhttp/metrics.go`

## Tests

```bash
go test -race -timeout 60s ./x/ops/...
go test -timeout 20s ./x/ops/...
go vet ./x/ops/...
```

## Docs Sync

Update `docs/modules/x-ops/README.md` if the public response shape changes or
if raw/bodyless health probe exceptions are intentionally retained.

## Done Definition

- `x/ops/healthhttp` has one documented success-response policy.
- JSON-body endpoints follow that policy consistently.
- Bodyless/raw probe exceptions, if any, are explicit and tested.
- The stable `health` root remains free of HTTP handler ownership.
- The listed validation commands pass.

## Outcome

