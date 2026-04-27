# Card 0574

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/README.md
- README.md
- docs/getting-started.md
Depends On: 2282, 2283

Goal:
Expose scenario scaffold profiles for the main user paths without changing Plumego's canonical application model.

Scope:
- Add or document scaffold profiles for REST API, tenant API, gateway/realtime, AI service, and ops-oriented services.
- Reuse the canonical app structure and explicit route/middleware wiring in generated output.
- Add tests that generated profiles include the expected capability imports and do not hide setup behind globals.
- Update CLI and onboarding docs with the available profile names.

Non-goals:
- Do not generate production secrets or provider credentials.
- Do not add external dependencies to generated templates.
- Do not make any `x/*` profile the default template.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`
- `README.md`
- `docs/getting-started.md`

Tests:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/0574-cmd-plumego-scenario-scaffold-profiles.md`

Docs Sync:
- Required because CLI scaffold behavior and onboarding guidance change.

Done Definition:
- The CLI exposes scenario profiles matching the documented scenario entrypoints.
- Generated templates preserve constructor-based, explicit wiring.
- Existing canonical and api templates remain compatible.

Outcome:
- Added scenario scaffold profiles: `rest-api`, `tenant-api`, `gateway`,
  `realtime`, `ai-service`, and `ops-service`.
- Kept scenario profiles on the canonical scaffold shape and added explicit
  `internal/scenario/profile.go` capability markers instead of hidden globals.
- Added scaffold tests for expected capability imports and no hidden `init`
  registration.
- Updated CLI, README, and getting-started docs with the new profile names.

Validations:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
