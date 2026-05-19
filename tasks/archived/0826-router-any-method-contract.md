# Card 0826

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/router.go, router/dispatch.go, core/routing.go, core/routing_test.go, docs/modules/router/README.md
Depends On: 0725-router-canonical-route-paths

Goal:
Make the ANY route method contract explicit and shared between `core` and
`router`.

Scope:
- Add a single exported router constant for the reserved ANY route method.
- Use that constant from `core.App.Any`.
- Document that ANY is a router fallback sentinel, not a custom exact HTTP
  method.
- Keep existing fallback behavior and route snapshots compatible.

Non-goals:
- Adding host routing.
- Changing standard HTTP method helpers.
- Supporting exact custom method `ANY` separately from fallback semantics.

Files:
- router/router.go
- router/dispatch.go
- core/routing.go
- core/routing_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/... ./core/...
- go test -race -timeout 60s ./router/... ./core/...
- go vet ./router/... ./core/...

Docs Sync:
- Required.

Done Definition:
- `core` no longer carries a duplicate ANY sentinel.
- Direct router users have a documented ANY constant and contract.
- Router/core tests, race tests, and vet pass.

Outcome:
- Added `router.MethodAny` as the shared reserved fallback method sentinel.
- Updated `core.App.Any` and core/router tests to use the router-owned
  constant instead of duplicate private sentinel strings.
- Documented that ANY is fallback routing semantics, not a separate exact
  custom HTTP method, and synchronized router public entrypoint docs.

Validation:
- `go test -timeout 20s ./router/... ./core/...`
- `go test -race -timeout 60s ./router/... ./core/...`
- `go vet ./router/... ./core/...`
