# Card 2263

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P1
State: active
Primary Module: x/gateway
Owned Files:
- x/gateway/module.yaml
- docs/modules/x-gateway/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2262

Goal:
Evaluate `x/gateway` for `beta` readiness while preserving explicit edge-transport wiring.

Scope:
- Check gateway, backend pool, proxy, rewrite, transform, cache, and protocol middleware entrypoints against beta criteria.
- Verify that nil-safe registration and invalid target behavior are documented and tested.
- If criteria are met, update status and docs.
- If criteria are not met, record the blocker by gateway subpackage.

Non-goals:
- Do not add new gateway protocols.
- Do not couple discovery selection to gateway defaults.
- Do not move edge policy into stable `router` or `middleware`.

Files:
- `x/gateway/module.yaml`
- `docs/modules/x-gateway/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when status or blocker language changes.

Done Definition:
- `x/gateway` has a policy-compliant promotion decision.
- Any remaining blocker is actionable and scoped to a gateway-owned package.

Outcome:
