# Card 2261

Milestone:
Recipe: specs/change-recipes/tenant-policy-change.yaml
Priority: P1
State: active
Primary Module: x/tenant
Owned Files:
- x/tenant/module.yaml
- docs/modules/x-tenant/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2260

Goal:
Evaluate `x/tenant` for `beta` readiness without weakening tenant fail-closed or stable-root boundary guarantees.

Scope:
- Check resolution, policy, quota, rate-limit, session, and tenant-aware store entrypoints against beta criteria.
- Verify that documented tenant isolation and fail-closed behavior have matching tests.
- If criteria are met, update status and docs.
- If criteria are not met, record blockers by subpackage.

Non-goals:
- Do not move tenant behavior into stable `middleware` or `store`.
- Do not add application-specific tenant CRUD or onboarding.
- Do not change tenant policy semantics.

Files:
- `x/tenant/module.yaml`
- `docs/modules/x-tenant/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when status or blocker language changes.

Done Definition:
- `x/tenant` has a policy-compliant promotion decision.
- Any remaining blocker identifies the exact tenant subpackage and missing evidence.

Outcome:
