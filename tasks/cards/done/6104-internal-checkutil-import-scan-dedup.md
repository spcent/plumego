# Card 6104

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: internal/checks/checkutil
Owned Files: internal/checks/checkutil/checkutil.go
Depends On: tasks/cards/done/6103-internal-validator-regex-hygiene.md

Goal:
Reduce duplicated Go import scanning logic in control-plane checks.

Scope:
- Extract one helper for walking Go files and reporting decoded imports.
- Use the helper in dependency-rule import checks and reference `x/*` import checks.
- Preserve existing violation strings and test behavior.

Non-goals:
- Do not replace the lightweight YAML scanners.
- Do not change check output format.
- Do not widen dependency-rule enforcement.

Files:
- internal/checks/checkutil/checkutil.go

Tests:
- go test ./internal/checks/checkutil
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/agent-workflow

Docs Sync:
- None; internal refactor only.

Done Definition:
- Import parsing error paths remain contextual.
- Existing checkutil tests pass.
- Boundary checks continue to pass.

Outcome:
- Added a shared Go import walker for control-plane checks.
- Reused it in disallowed import checks and reference `x/*` import checks.
- Preserved violation formatting and missing-directory behavior.
- Validation: `go test ./internal/checks/checkutil`; `go run ./internal/checks/dependency-rules`; `go run ./internal/checks/agent-workflow`; `go test ./internal/...`.
