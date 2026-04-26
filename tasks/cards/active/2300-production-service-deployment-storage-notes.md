# Card 2300

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: reference
Owned Files:
- reference/production-service/README.md
- reference/production-service/internal/app/routes.go
- reference/production-service/internal/config/config.go
- docs/modules/x-ops/README.md
- tasks/cards/active/README.md
Depends On: 2292

Goal:
Deepen production-service guidance for deployment, storage replacement, and
protected operations without hiding wiring behind a bundle.

Scope:
- Add deployment and durable-storage replacement notes.
- Add route/config output that exposes the reference storage/auth policy.
- Keep all storage app-local and standard-library-only.

Non-goals:
- Do not add external database dependencies.
- Do not mount devtools.
- Do not introduce hidden production bundles.

Files:
- `reference/production-service/README.md`
- `reference/production-service/internal/app/routes.go`
- `reference/production-service/internal/config/config.go`
- `docs/modules/x-ops/README.md`
- `tasks/cards/active/README.md`

Tests:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2300-production-service-deployment-storage-notes.md`

Docs Sync:
- Required because production reference guidance changes.

Done Definition:
- Production reference explains deployment, secret, storage, and ops exposure
  decisions using implemented behavior only.

Outcome:
