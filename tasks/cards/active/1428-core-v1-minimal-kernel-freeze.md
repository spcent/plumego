# Card 1428

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: core
Owned Files:
- core/*
Depends On:
- 1421
- 1423
- 1424

Goal:
- Freeze `core` as a minimal application kernel after the transport, routing,
  and contract surfaces are normalized.

Scope:
- Enumerate lifecycle, HTTP assembly, route attachment, and middleware
  convenience APIs before editing.
- Remove convenience wrappers that duplicate app wiring responsibilities.
- Keep lifecycle order explicit and test-covered.
- Keep `core` free of routing internals, middleware catalogs, feature
  registries, and `x/*` imports.
- Update tests and docs for the final v1 kernel API.

Non-goals:
- Do not add plugin discovery, feature catalogs, or hidden registration.
- Do not change handler compatibility with `net/http`.
- Do not import extension packages.

Files:
- core/app.go
- core/lifecycle.go
- core/http_handler.go
- core/routing.go
- core/middleware.go
- core/*_test.go

Tests:
- go test -timeout 20s ./core
- go vet ./core
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update core docs and reference snippets if public construction or lifecycle
  APIs change.

Done Definition:
- `core` exposes the minimal v1 kernel API.
- Removed wrappers have no remaining callers.
- Core tests and boundary checks pass.

Outcome:

