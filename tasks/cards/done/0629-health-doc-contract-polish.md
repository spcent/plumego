# Card 0629: Health Doc Contract Polish

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: health
Owned Files:
- `health/core.go`
- `health/readiness.go`
- `health/module.yaml`
- `docs/modules/health/README.md`
Depends On: 5404

Goal:
Align package comments, module metadata, and module docs with the actual stable
health model semantics.

Scope:
- Clarify that `HealthStatus` can describe component or aggregate state.
- Clarify `ReadinessStatus.Components` as a component-name readiness map.
- Document readiness semantics for healthy, degraded, and unhealthy states.
- Keep the boundary language transport-free and manager-free.

Non-goals:
- Do not change DTO fields, JSON tags, or public APIs.
- Do not move orchestration or HTTP behavior into `health`.
- Do not update unrelated module docs.

Files:
- `health/core.go`
- `health/readiness.go`
- `health/module.yaml`
- `docs/modules/health/README.md`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
This card is the docs sync for the stable health model wording.

Done Definition:
- Code comments, module metadata, and docs describe the same health semantics.
- The listed validation commands pass.

Outcome:
- Clarified `HealthStatus`, `ComponentChecker`, `ComponentHealth`, and
  `ReadinessStatus` comments.
- Updated `health/module.yaml` summary, responsibilities, and review checklist
  wording without expanding the checklist beyond manifest limits.
- Added module docs for readiness semantics, aggregate health use, component
  readiness booleans, and standard `time.Duration` JSON encoding.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
  - `go run ./internal/checks/module-manifests`
