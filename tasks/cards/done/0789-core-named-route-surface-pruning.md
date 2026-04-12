# Card 0789

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/routing.go`
- `core/routing_test.go`
- `docs/modules/core/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
Depends On:
- `0787-core-router-policy-sync-side-effect-removal.md`

Goal:
- Prune redundant named-route shorthand helpers so `core` keeps one clear named
  route registration path.

Problem:
- `core` exposes `AddRouteWithName(...)` plus `GetNamed(...)`,
  `PostNamed(...)`, `PutNamed(...)`, `DeleteNamed(...)`, `PatchNamed(...)`, and
  `AnyNamed(...)`.
- The per-method named helpers are only convenience aliases over the generic
  named route API and currently appear only in tests.
- This leaves the kernel with multiple public ways to express the same named
  route registration behavior.

Scope:
- Keep one canonical named-route registration path.
- Remove redundant named-route shorthands.
- Update tests and docs to the reduced route surface.

Non-goals:
- Do not redesign reverse URL lookup.
- Do not remove unnamed `Get/Post/...` shorthands unless the chosen canonical
  named path requires it.
- Do not add new route helper families.

Files:
- `core/routing.go`
- `core/routing_test.go`
- `docs/modules/core/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Keep route-registration guidance aligned with the reduced named-route API.

Done Definition:
- `core` exposes one canonical named-route registration path.
- Redundant `*Named(...)` helper aliases are removed.
- Tests and docs no longer rely on deleted named-route shorthands.
