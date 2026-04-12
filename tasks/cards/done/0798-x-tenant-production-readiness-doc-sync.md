# Card 0798

Priority: P2
State: done
Primary Module: x/tenant
Owned Files:
- `docs/modules/x-tenant/README.md`
- `docs/architecture/X_TENANT_BLUEPRINT.md`
- `x/tenant/module.yaml`
- `docs/ROADMAP.md`
Depends On:
- `0795-x-tenant-resolution-examples.md`
- `0796-x-tenant-quota-retry-after-coverage.md`
- `0797-x-tenant-policy-and-isolation-coverage.md`

Goal:
- Sync the tenant docs surface with the implemented production-readiness examples and coverage.

Scope:
- Update the tenant module primer, blueprint, manifest metadata, and roadmap wording to reflect implemented resolution examples and coverage depth.
- Remove any stale or underspecified wording left behind by earlier tenant cards.
- Keep the docs grounded in implemented behavior only.

Non-goals:
- Do not add new runtime behavior in this card.
- Do not promote `x/tenant` to a stable root or a GA extension.
- Do not rewrite unrelated roadmap phases.

Files:
- `docs/modules/x-tenant/README.md`
- `docs/architecture/X_TENANT_BLUEPRINT.md`
- `x/tenant/module.yaml`
- `docs/ROADMAP.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync:
- This card is the tenant sync pass; keep module docs, blueprint, manifest, and roadmap wording aligned with implemented behavior.

Done Definition:
- Tenant docs and manifest metadata describe the same implemented readiness surface.
- No stale wording suggests undocumented behavior or stronger guarantees than the tests support.
- The roadmap reflects what actually landed.

Outcome:
- Synced the tenant module primer, blueprint, manifest wording, and roadmap language with the example-backed resolution flows and the new quota/policy/ratelimit coverage.
- Updated Phase 9 in `docs/ROADMAP.md` from stale planned-only wording to an in-progress state that reflects what is now implemented and what remains.
- Kept the docs grounded in implemented behavior only; the only validation issue was a manifest schema limit on `agent_hints`, which was corrected without changing runtime behavior.

Validation Run:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`
