# Card 0597

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/README.md
- docs/getting-started.md
- tasks/cards/active/README.md
Depends On: 2299

Goal:
Make `ai-service` and `ops-service` scaffold profiles runnable instead of only
capability markers.

Scope:
- Add offline AI provider/session/tool routes for the `ai-service` template.
- Add protected ops route and metrics/health/admin boundary examples for the
  `ops-service` template.
- Update CLI and getting-started docs to describe implemented routes.

Non-goals:
- Do not require live AI provider credentials.
- Do not mount devtools by default.
- Do not add new runtime dependencies.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `tasks/cards/active/README.md`

Tests:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/0597-runnable-ai-ops-scaffold-profiles.md`

Docs Sync:
- Required because generated scaffold behavior changes.

Done Definition:
- `ai-service` and `ops-service` generated projects expose runnable routes that
  demonstrate the selected capability families with explicit wiring.

Outcome:
- Added runnable `ai-service` scaffold routes with an offline mock provider,
  memory session manager, echo tool registry, and `/ai/demo` response.
- Added runnable `ops-service` scaffold routes with protected
  observability-backed `/ops/metrics` and protected `/ops/admin` boundary
  summary using `x/ops` DTOs.
- Updated CLI and getting-started docs to describe the implemented AI and ops
  scenario routes.

Validations:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/0597-runnable-ai-ops-scaffold-profiles.md`
