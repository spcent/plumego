# Card 0002

Priority: P1

Goal:
- Add a repo check that ensures `specs/change-recipes/*` remains discoverable from canonical repo metadata.

Scope:
- new workflow-oriented check under `internal/checks`
- `specs/repo.yaml`

Non-goals:
- Do not validate recipe content quality in this card.
- Do not change task-card format.

Files:
- `internal/checks/agent-workflow/main.go`
- `internal/checks/checkutil/checkutil.go`
- `specs/repo.yaml`
- `AGENTS.md`
- `CLAUDE.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go test ./internal/checks/checkutil/...`

Docs Sync:
- Sync required quality gates in `AGENTS.md` and `CLAUDE.md` if a new check is introduced.

Done Definition:
- Removing recipe discoverability from canonical repo metadata causes the new check to fail.
- The check is cheap enough for routine use.
- The required gate list stays synchronized with the implemented checks.
