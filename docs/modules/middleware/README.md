# middleware

## Purpose

`middleware` contains transport-only HTTP middleware.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- adding request/response wrappers
- enforcing ordering-sensitive transport behavior
- instrumenting requests without owning business policy

## Do not use this module for

- business validation
- tenant policy catalogs
- tenant resolution or tenant quota enforcement
- service construction

## First files to read

- `middleware/module.yaml`
- `docs/CANONICAL_STYLE_GUIDE.md`
- the target package under `middleware/*`

## Canonical change shape

- implement `func(http.Handler) http.Handler`
- add ordering and error-path tests
- keep side effects explicit and local
- keep tenant-aware policy, resolution, and quota behavior in `x/tenant`
- keep auth and security-header transport adapters here, on top of `security/*` primitives

## Boundary with observability

- stable `middleware/*` owns transport-only observability primitives such as request IDs, tracing hooks, access logging, and HTTP metrics
- `x/observability` owns broader adapter, export, and integration wiring
- do not turn stable `middleware` into an observability catch-all catalog
