# x/tenant

## Purpose

`x/tenant` is the extension boundary for multi-tenant policy, quota, rate limit, resolution, and tenant-aware stores.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is tenant policy or isolation work
- the change is tenant-aware by design

## Do not use this module for

- stable middleware defaults
- stable store defaults
- generic request middleware unrelated to tenant semantics

## Main entrypoints

- `x/tenant/resolve` — tenant identity extraction and request-scoped resolution
- `x/tenant/policy` — allow/deny decisions and policy evaluation
- `x/tenant/quota` — usage budgeting and quota enforcement
- `x/tenant/ratelimit` — tenant-scoped rate limiting
- `x/tenant/config` — tenant configuration and management helpers
- `x/tenant/store/*` — tenant-aware cache and database adapters

## First files to read

- `x/tenant/module.yaml`
- `docs/architecture/X_TENANT_BLUEPRINT.md`
- the owning subpackage under `x/tenant/*`
- `AGENTS.md` tenant boundary rules

## Boundary rules

- keep tenant-aware logic out of stable `middleware` and stable `store`
- fail closed on resolution, policy, quota, and validation errors
- keep reference apps tenant-agnostic by default
- keep application-specific tenant CRUD outside this module

## Canonical change shapes

- resolution work starts in `x/tenant/resolve`
- deny-path, quota, and policy work starts in `x/tenant/core`, `policy`, `quota`, or `ratelimit`
- tenant-aware persistence work starts in `x/tenant/store/*`, not in stable `store/*`

## Validation focus

- `go test -race -timeout 60s ./x/tenant/...`
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`
- add negative-path coverage for isolation, quota exhaustion, and policy-deny behavior when changing public flows
