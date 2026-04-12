# Card 0753

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/routing.go`
- `core/routing_test.go`
- `reference/standard-service/internal/app/routes.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `docs/CANONICAL_STYLE_GUIDE.md`
Depends On:
- `0752-core-lifecycle-surface-convergence.md`

Goal:
- Make route-registration failures visible at wiring time and stop treating the
  silent log-only helpers as the canonical `core` API surface.

Problem:
- `AddRoute` and `AddRouteWithName` return explicit errors.
- `Get`, `Post`, `Put`, `Delete`, `Patch`, `Any`, `Handle`, and the named
  variants swallow registration failures and only log them through
  `registerRoute` / `registerNamedRoute`.
- That gives `core` two parallel route-registration models: one strict and one
  silent. The reference app and scaffold still use the silent path, so the
  canonical examples do not surface duplicate-route or post-freeze failures.

Scope:
- Move canonical first-party wiring to the strict error-returning registration
  path.
- Keep method-per-line route readability in examples; do not regress to hidden
  registration loops or reflection.
- Remove the silent log-only registration helpers so `core` exposes a single
  strict route-registration model.

Non-goals:
- Do not redesign `router` route matching or naming.
- Do not add route groups or plugin-style registries in `core`.
- Do not add a second “convenience” registration path back in.

Files:
- `core/routing.go`
- `core/routing_test.go`
- `reference/standard-service/internal/app/routes.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `docs/CANONICAL_STYLE_GUIDE.md`

Tests:
- Add coverage that duplicate or invalid route registration errors are surfaced
  in the canonical wiring path.
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./cmd/plumego/internal/scaffold/...`

Docs Sync:
- Update the style guide so canonical route wiring uses the strict registration
  path while preserving one method + path + handler per line.

Done Definition:
- The silent log-only registration helpers are removed.
- The reference app and scaffold use the strict registration path.
- `core` route-registration tests cover strict failure surfacing.
- Canonical docs describe a single preferred registration model.

Outcome:
- Removed the silent log-only registration helpers from `core/routing.go`.
- Changed `Get` / `Post` / `Put` / `Delete` / `Patch` / `Any` / `Handle` /
  `HandleFunc` and named variants to return explicit registration errors.
- Updated `core` tests to treat method helpers as the canonical strict wiring
  path and added coverage that duplicate and post-start registration failures
  surface through those helpers.
- Migrated first-party route wiring in the reference apps, dev dashboard, and
  scaffold templates to handle registration errors explicitly.
- Updated `docs/CANONICAL_STYLE_GUIDE.md` so canonical route wiring returns
  registration errors while keeping one method + path + handler per line.
- Validation:
  - `gofmt -w core/routing.go core/test_helpers_test.go core/lifecycle_test.go core/app_test.go core/options_test.go core/routing_test.go cmd/plumego/internal/devserver/dashboard.go cmd/plumego/internal/scaffold/scaffold.go reference/standard-service/internal/app/routes.go reference/with-gateway/internal/app/routes.go reference/with-messaging/internal/app/routes.go reference/with-webhook/internal/app/routes.go reference/with-websocket/internal/app/routes.go`
  - `go test -race -timeout 60s ./core/...`
  - `go test -timeout 20s ./reference/...`
  - `go test -timeout 20s ./internal/scaffold/... ./internal/devserver/...` (run from `cmd/plumego/`)
  - `go vet ./core/... ./reference/...`
  - `go vet ./internal/scaffold/... ./internal/devserver/...` (run from `cmd/plumego/`)
