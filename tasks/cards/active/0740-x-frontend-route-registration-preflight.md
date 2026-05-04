# Card 0740: x/frontend Route Registration Preflight

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0739

Goal:
Reduce partial route registration risk by making mount route plans explicit and
preflighting known router snapshots before mutating the router.

Scope:
- Expose or internally use an ordered route plan for each mount.
- Before registering into a registrar that exposes existing route snapshots,
  detect duplicate mount routes and fail before adding any new route.
- Keep the minimal `Registrar` contract compatible with `router.Router` and
  `core.App`.
- Add coverage for duplicate root/prefix route preflight and custom registrar
  fallback behavior.

Non-goals:
- Do not add route removal APIs to `router`.
- Do not change router matching behavior.
- Do not require all custom registrars to expose snapshots.

Files:
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document route registration order and the preflight behavior for snapshot-capable
registrars.

Done Definition:
- Duplicate frontend mount routes are rejected before partial mutation when the
  registrar exposes route snapshots.
- `router.Router` and `core.App` remain compatible registration targets.
- Custom AddRoute-only registrars keep current best-effort behavior.
- The listed validation commands pass.
