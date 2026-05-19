# Card 1432

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`
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
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`

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
- Recorded CLI and onboarding evidence in `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`.
- Updated `docs/release/v1.0.0-rc.1.md` with source-checkout CLI validation
  evidence and the tagged-install limitation.
- Validation passed:
  - `cd cmd/plumego && go test -timeout 20s ./...`
  - `cd cmd/plumego && go vet ./...`
  - `cd cmd/plumego && go run . new --template canonical --dry-run trust-check`
- Tagged install smoke was attempted:
  - first attempt failed because sandboxed execution could not write the default
    Go sumdb cache
  - retry with `GOPATH`, `GOBIN`, `GOCACHE`, and `GOMODCACHE` under
    `/private/tmp` failed with `module github.com/spcent/plumego@v1.0.0-rc.1
    found, but does not contain package github.com/spcent/plumego/cmd/plumego`
- Result: keep CLI install guidance source-checkout based for this rc.
