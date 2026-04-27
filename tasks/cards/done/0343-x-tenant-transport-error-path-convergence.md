# Card 0343: X Tenant Transport Error Path Convergence

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/tenant
Depends On: —

## Goal

Remove the extra tenant-specific error write wrapper so `x/tenant` denial paths
use one canonical transport error path while keeping tenant header helpers
explicit and reusable.

## Problem

- `x/tenant/transport/helpers.go` exports `WriteError`, but that function is
  only a thin wrapper over `middleware.WriteTransportError`.
- `x/tenant/resolve`, `x/tenant/policy`, `x/tenant/quota`, and
  `x/tenant/ratelimit` all depend on that wrapper, while `x/tenant/session`
  already writes errors through `contract.WriteError` directly.
- This leaves the same extension family with two different error-write idioms
  and violates the repository rule that each layer should converge on one
  canonical error-construction path.

## Scope

- Enumerate and migrate every in-repo `transport.WriteError` caller.
- Remove the exported `WriteError` helper from `x/tenant/transport` once the
  last caller is migrated.
- Keep `Retry-After`, rate-limit, and quota header helpers in
  `x/tenant/transport`.
- Update focused tests and package comments that still teach the removed helper.

## Non-Goals

- Do not change tenant quota, rate-limit, or policy semantics.
- Do not change header names or remaining-budget behavior.
- Do not move tenant logic into stable `middleware`.

## Files

- `x/tenant/transport/helpers.go`
- `x/tenant/transport/doc.go`
- `x/tenant/resolve/middleware.go`
- `x/tenant/policy/middleware.go`
- `x/tenant/quota/middleware.go`
- `x/tenant/ratelimit/middleware.go`
- related `x/tenant/*_test.go`

## Tests

```bash
rg -n 'tenanttransport\\.WriteError|transport\\.WriteError\\(' x/tenant . -g '*.go'
go test -timeout 20s ./x/tenant/...
go vet ./x/tenant/...
```

## Docs Sync

- `docs/modules/x-tenant/README.md` only if it still presents
  `transport.WriteError` as part of the canonical tenant middleware surface

## Done Definition

- `x/tenant` no longer exports or uses a second transport error write helper.
- Tenant denial paths share one canonical error-construction/write path.
- Header propagation helpers remain explicit and test-backed.
- Grep for the removed `WriteError` helper is empty outside intentional history.

## Outcome

- Removed `WriteError` from `x/tenant/transport`, leaving that package focused
  on header names, retry-after propagation, and remaining-budget headers.
- Migrated `resolve`, `policy`, `quota`, and `ratelimit` deny paths to direct
  `contract.NewErrorBuilder` + `contract.WriteError` calls.
- Added/kept focused middleware tests that assert canonical tenant error codes
  still surface on invalid tenant IDs, missing tenant IDs, quota rejections,
  policy denials, and rate-limit rejections.

## Validation Run

```bash
rg -n 'tenanttransport\.WriteError|transport\.WriteError\(' x/tenant . -g '*.go'
go test -timeout 20s ./x/tenant/...
go vet ./x/tenant/...
```
