# Card 0384

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/main.go`
- `reference/workerfleet/main_test.go`
- `reference/workerfleet/README.md`
- `tasks/cards/done/0384-workerfleet-service-entrypoint.md`
Depends On:
- `tasks/cards/done/0383-workerfleet-grafana-dashboard-docs.md`
Blocked By: —

Goal:
- Add the workerfleet service entrypoint so the reference app can be started as an HTTP service.
- Wire existing bootstrap, routes, and Prometheus metrics into a `net/http` compatible Plumego server.

Scope:
- Add `package main` under `reference/workerfleet`.
- Load workerfleet app config from environment.
- Load HTTP server address and shutdown timeout from environment.
- Register workerfleet API routes and `/metrics`.
- Start the Plumego HTTP server with graceful signal shutdown.
- Document local run and build commands.

Non-goals:
- Do not add Kubernetes sync or alert notification background loops in this card.
- Do not add new external dependencies.
- Do not modify repository root `go.mod`.
- Do not change workerfleet API contracts.

Files:
- `reference/workerfleet/main.go`
- `reference/workerfleet/main_test.go`
- `reference/workerfleet/README.md`

Tests:
- `cd reference/workerfleet && go test ./...`
- `cd reference/workerfleet && go build .`

Docs Sync:
- Update `reference/workerfleet/README.md` with startup commands and entrypoint environment variables.

Done Definition:
- `cd reference/workerfleet && go run .` starts an HTTP service when required store config is present.
- `/metrics` is wired through the existing metrics exporter.
- Shutdown handles SIGINT/SIGTERM gracefully.

Outcome:
- Added `reference/workerfleet/main.go` with environment config loading, Plumego route wiring, `/metrics`, server startup, and SIGINT/SIGTERM graceful shutdown.
- Added entrypoint config tests and README startup documentation.
- Kubernetes sync and alert notification loops remain out of scope for this card.
