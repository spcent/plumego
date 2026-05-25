# Card 1508

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/module.yaml`
- `x/openapi/module.yaml`
- `x/validate/module.yaml`
- `docs/modules/x-frontend/README.md`
- `docs/modules/x-openapi/README.md`
Depends On: 1503

Goal:
- Add missing `public_entrypoints` to `x/frontend`, `x/openapi`, and
  `x/validate`.
- Keep the documented public surface aligned with the chosen manifest
  convention.

Scope:
- Update the three manifests above and only the matching docs sections needed
  to explain their public entrypoint inventory.

Non-goals:
- Do not change runtime behavior in any of the three modules.
- Do not add new exported symbols.
- Do not widen this card into `x/rpc`, `x/gateway`, or `x/observability`.

Files:
- `x/frontend/module.yaml`
- `x/openapi/module.yaml`
- `x/validate/module.yaml`
- `docs/modules/x-frontend/README.md`
- `docs/modules/x-openapi/README.md`

Acceptance Tests:
<!-- none; manifest/docs card -->

Tests:
- `go test -timeout 20s ./x/frontend/... ./x/openapi/... ./x/validate/...`
- `go run ./internal/checks/public-entrypoints-sync`

Docs Sync:
- `docs/modules/x-frontend/README.md`
- `docs/modules/x-openapi/README.md`

Validation:
- `go test -timeout 20s ./x/frontend/... ./x/openapi/... ./x/validate/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added concrete exported public entrypoint inventories to `x/frontend`,
  `x/openapi`, and `x/validate`.
- Removed the ghost `core/components/**` boundary rule from `x/frontend` while
  touching the manifest.
- Validation:
  - `go test -timeout 20s ./x/frontend/... ./x/openapi/... ./x/validate/...`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
