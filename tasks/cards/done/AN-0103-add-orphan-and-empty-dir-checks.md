# Card AN-0103

Priority: P0

Goal:
- Add a checker for orphaned extension roots and empty misleading directories.

Scope:
- orphaned extension root detection
- empty misleading directory detection
- failure messaging

Non-goals:
- Do not add package semantic rules yet.
- Do not redesign existing check runner structure in this card.

Files:
- `internal/checks/agent-workflow/main.go`
- `internal/checks/checkutil/checkutil.go`
- `internal/checks/checkutil/checkutil_test.go`
- `specs/repo.yaml`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep checker behavior aligned with the control-plane rules for extension registration and misleading directories.

Done Definition:
- New extension roots cannot silently exist outside the declared control plane.
- Empty misleading directories fail a repo check with actionable output.
- The new rule fits the existing agent-workflow validation flow.

Outcome:
- Added `layers.extension.paths` truth alignment for current top-level `x/*` roots in `specs/repo.yaml`.
- Added helper functions to parse declared extension roots, detect orphaned `x/*` roots, and flag empty misleading directories under stable roots and `x/*`.
- Extended `internal/checks/agent-workflow` to fail on those new violations.

Validation Run:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`
