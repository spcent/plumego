# Card 1432

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: blocked
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`
Depends On:
- 1431

Goal:
- Record CLI and onboarding smoke evidence for v1 release readiness.

Scope:
- Verify supported scaffold templates and generated-app commands.
- Confirm CLI help, README, and getting-started claims remain aligned.
- Record source checkout install guidance unless tagged install is proven.

Non-goals:
- Do not add scaffold templates.
- Do not change stable root API.
- Do not claim tagged CLI install support without smoke evidence.

Files:
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`

Tests:
- `cd cmd/plumego && go test -timeout 20s ./...`
- `cd cmd/plumego && go vet ./...`
- `cd cmd/plumego && go run . new --template canonical --dry-run trust-check`

Docs Sync:
- Required if CLI output, template names, or install claims differ from docs.

Done Definition:
- CLI smoke evidence is recorded.
- Onboarding docs match executable behavior.
- Any install blocker is explicit.

Outcome:
-
