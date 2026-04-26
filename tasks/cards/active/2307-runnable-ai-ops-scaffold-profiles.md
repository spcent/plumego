# Card 2307

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: active
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
- `scripts/check-spec tasks/cards/done/2307-runnable-ai-ops-scaffold-profiles.md`

Docs Sync:
- Required because generated scaffold behavior changes.

Done Definition:
- `ai-service` and `ops-service` generated projects expose runnable routes that
  demonstrate the selected capability families with explicit wiring.

Outcome:
