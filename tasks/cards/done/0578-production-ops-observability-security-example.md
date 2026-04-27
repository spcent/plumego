# Card 0578

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: reference
Owned Files:
- reference/production-service/README.md
- reference/production-service/internal/app/routes.go
- docs/modules/x-observability/README.md
- docs/modules/x-ops/README.md
- docs/modules/x-devtools/README.md
Depends On: 2281, 2283

Goal:
Strengthen the production reference with an explicit safe composition example for metrics, health, protected ops, and debug boundaries.

Scope:
- Show which routes are mounted by default in `reference/production-service`.
- Demonstrate protected ops/admin routing without exposing local debug surfaces.
- Document that metrics and health may be public or protected based on deployment policy, but devtools must remain local/protected and off by default.
- Align `x/observability`, `x/ops`, and `x/devtools` primers with the example.

Non-goals:
- Do not add auth secrets to the repository.
- Do not mount `x/devtools` by default.
- Do not create a hidden observability bundle.

Files:
- `reference/production-service/README.md`
- `reference/production-service/internal/app/routes.go`
- `docs/modules/x-observability/README.md`
- `docs/modules/x-ops/README.md`
- `docs/modules/x-devtools/README.md`

Tests:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/0578-production-ops-observability-security-example.md`

Docs Sync:
- Required because production route exposure guidance changes.

Done Definition:
- The production reference states which operational routes are mounted and why.
- Debug/devtools exposure remains explicitly opt-in and protected.
- Observability and ops primers match the concrete production example.

Outcome:
- Protected `reference/production-service` `/ops/metrics` with
  `middleware/auth` and `security/authn.StaticToken` using `OPS_TOKEN`.
- Documented public probe routes, protected ops metrics, fail-closed behavior
  when `OPS_TOKEN` is unset, and the no-default-devtools boundary.
- Aligned `x/observability`, `x/ops`, and `x/devtools` primers with the
  production reference route exposure model.

Validations:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/0578-production-ops-observability-security-example.md`
