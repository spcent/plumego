# Card 0948: Middleware Tenant Error Ownership Eviction

Priority: P0
State: active
Primary Module: middleware

## Goal

Remove tenant-specific transport error code ownership from stable `middleware` and move it to the owning tenant transport layer.

## Problem

- `middleware/error_registry.go` defines tenant-specific constants:
  - `CodeTenantRequired`
  - `CodeTenantInvalidID`
  - `CodeTenantPolicyDenied`
  - `CodeTenantQuotaExceeded`
  - `CodeTenantRateLimited`
- This conflicts with the stable boundary contract:
  - `middleware/module.yaml` says stable middleware must not own tenant policy, resolution, or quota behavior.
  - `docs/modules/middleware/README.md` says tenant-aware policy and quota belong in `x/tenant`.
- The boundary leak is not theoretical:
  - `x/tenant/transport/helpers.go` aliases these stable middleware constants back into the tenant package.
  - `middleware` conformance tests assert tenant error codes owned by the stable root.
- The stable package should keep only transport-generic error writing and transport-generic codes such as `server_busy`, `server_queue_timeout`, and `upstream_failed`.

## Scope

- Move tenant transport error code ownership out of `middleware/error_registry.go` into the owning `x/tenant/transport` package.
- Keep `middleware.WriteTransportError(...)` as the canonical generic write helper unless that helper also proves tenant-coupled.
- Update `x/tenant` helpers and middleware conformance tests so stable middleware no longer exports tenant-specific codes.
- Confirm that extension packages still return the same wire-format codes after the ownership move.

## Non-Goals

- Do not change non-tenant middleware transport codes.
- Do not redesign tenant resolution or quota behavior.
- Do not move generic error writing into `contract` unless the cleanup clearly requires it.

## Files

- `middleware/error_registry.go`
- `middleware/conformance_test.go`
- `middleware/conformance/runtime_invariants_test.go`
- `x/tenant/transport/helpers.go`

## Tests

- `go test -timeout 20s ./middleware/... ./x/tenant/...`
- `go test -race -timeout 60s ./middleware/... ./x/tenant/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

- Update module docs only if the final ownership or helper entrypoint changes in a way not already covered by the existing middleware and tenant READMEs.

## Done Definition

- Stable `middleware` no longer exports tenant-specific error-code constants.
- `x/tenant` owns tenant transport error identifiers directly instead of aliasing back into stable middleware.
- Middleware conformance tests stop encoding tenant catalog ownership into the stable root.
- Tenant transport responses preserve their existing structured error shape and codes.

## Outcome

- Removed tenant-specific error code constants from `middleware/error_registry.go`.
- Moved tenant transport code ownership to `x/tenant/transport/helpers.go`.
- Updated middleware conformance tests to assert tenant error codes through the tenant transport package instead of the stable middleware root.
- Kept `middleware.WriteTransportError(...)` as the generic stable transport write helper.

## Validation Run

```bash
gofmt -w middleware/error_registry.go middleware/conformance_test.go middleware/conformance/runtime_invariants_test.go x/tenant/transport/helpers.go
rg -n 'CodeTenantRequired|CodeTenantInvalidID|CodeTenantPolicyDenied|CodeTenantQuotaExceeded|CodeTenantRateLimited' middleware x/tenant
go test -timeout 20s ./middleware/... ./x/tenant/...
go run ./internal/checks/dependency-rules
```
