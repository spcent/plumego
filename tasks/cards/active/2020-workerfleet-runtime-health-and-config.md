# Card 2020

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/config_test.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/internal/handler/health.go`
- `reference/workerfleet/main.go`
Depends On:
- `tasks/cards/done/2019-workerfleet-service-entrypoint.md`
Blocked By: —

Goal:
- Add production-facing runtime configuration and health/readiness endpoints for workerfleet.
- Make operators able to verify HTTP liveness and storage readiness before enabling background loops.

Scope:
- Add app config fields for Kubernetes sync interval, status sweep interval, alert evaluation interval, notifier delivery timeout, and runtime feature enable flags.
- Add `GET /healthz` and `GET /readyz`.
- Make readiness verify the configured store is available through existing runtime wiring.
- Keep config parsing environment-driven and fail closed on invalid durations.

Non-goals:
- Do not start Kubernetes sync, status sweeper, alert evaluation, or notification loops in this card.
- Do not add Kubernetes or notifier credentials.
- Do not change worker API contracts.

Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/config_test.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/internal/handler/health.go`
- `reference/workerfleet/main.go`

Tests:
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./internal/handler/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update workerfleet README only if new environment variables are added.

Done Definition:
- `/healthz` returns liveness without external dependency checks.
- `/readyz` returns readiness based on initialized runtime dependencies.
- Invalid runtime interval config fails before the server starts.

