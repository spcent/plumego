# Card 2306

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: active
Primary Module: reference/production-service
Owned Files:
- reference/production-service/internal/app/routes.go
- reference/production-service/internal/config/config.go
- reference/production-service/internal/app/app_test.go
- reference/production-service/README.md
- docs/modules/x-ops/README.md
Depends On: 2300

Goal:
Deepen `reference/production-service` into a production vertical slice with a
durable profile-store example, config matrix, protected ops/admin/debug policy,
and endpoint smoke coverage.

Scope:
- Add a standard-library durable profile store option.
- Add endpoint smoke tests for startup wiring, health, metrics, and protected
  profile behavior.
- Document config and protected surface policy.

Non-goals:
- Do not add external database dependencies.
- Do not mount `x/devtools` by default.
- Do not make production-service a hidden framework bundle.

Files:
- `reference/production-service/internal/app/routes.go`
- `reference/production-service/internal/config/config.go`
- `reference/production-service/internal/app/app_test.go`
- `reference/production-service/README.md`
- `docs/modules/x-ops/README.md`

Tests:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2306-production-service-durable-profile-smoke.md`

Docs Sync:
- Required because production reference behavior and config change.

Done Definition:
- Production reference demonstrates a durable storage replacement path and
  smoke-tested protected operational surface without adding external deps.

Outcome:
