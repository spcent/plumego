# Card 0121

Milestone:
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
  - docs/modules/middleware/README.md
  - middleware/middleware.go
  - cmd/plumego/internal/scaffold/scaffold.go
  - reference/production-service/internal/app/app.go
  - docs/getting-started.md
Depends On: 0120

Goal:
Align documented and generated middleware ordering so recovery can catch
downstream middleware panics while request IDs remain available for normal
request telemetry.

Scope:
- Pick one canonical production order and apply it consistently.
- Fix examples in module docs, package docs, getting-started docs, scaffolds, and production reference.
- Keep changes to wiring order and comments only.

Non-goals:
- Do not change middleware implementation.
- Do not add a hidden production bundle.
- Do not alter application-specific route behavior.

Files:
- `docs/modules/middleware/README.md`
- `middleware/middleware.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `reference/production-service/internal/app/app.go`
- `docs/getting-started.md`

Tests:
- `go test -timeout 20s ./cmd/plumego/internal/scaffold ./reference/production-service/...`
- `go test -timeout 20s ./middleware/...`
- `go vet ./...`

Docs Sync:
- Required for all edited docs and generated examples.

Done Definition:
- No conflicting recommended ordering remains in middleware docs/examples/scaffold.
- Targeted tests and vet pass.

Outcome:

