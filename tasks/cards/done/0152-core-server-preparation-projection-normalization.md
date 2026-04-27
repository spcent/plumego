# Card 0152

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/config.go`
- `core/lifecycle_test.go`
- `core/introspection_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Normalize server preparation and snapshot projection so `core` stops manually
  remapping the same server config in multiple places.

Problem:
- `core/lifecycle.go` still keeps a private `serverRuntimeConfig` projection
  and an unused `setupServer()` wrapper.
- `core/http_handler.go` maps that projection into `http.Server`.
- `core/introspection.go` separately remaps the same config fields into
  `RuntimeSnapshot`.
- This leaves the same kernel-owned server settings copied across multiple
  structs/functions, making drift easy and dead helpers harder to spot.

Scope:
- Remove dead lifecycle helpers that are no longer called.
- Converge `http.Server` preparation and runtime snapshot export onto one
  internal typed projection/helper path.
- Keep current server behaviour and defaults unchanged.

Non-goals:
- Do not redesign TLS loading semantics in this card.
- Do not change lifecycle ordering.
- Do not widen the `core` public API.

Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/config.go`
- `core/lifecycle_test.go`
- `core/introspection_test.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Keep the core primer aligned on one internal server-preparation model.

Done Definition:
- `core` no longer carries dead private preparation helpers.
- Server config fields are projected through one internal path instead of
  repeated manual copies.
- Snapshot and preparation code stay aligned through shared normalization.

Outcome:
- Removed the dead private `setupServer()` wrapper and the duplicate
  `serverRuntimeConfig` projection from `core/lifecycle.go`.
- Added a single shared `projectRuntimeSnapshot(...)` helper that now feeds both
  prepared `http.Server` construction and `RuntimeSnapshot()` export.
- Updated the core primer to describe the single internal preparation
  projection, keeping server preparation and snapshot behavior aligned.
