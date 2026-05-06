# Card 0760

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego check
Owned Files: cmd/plumego/commands/check.go, cmd/plumego/commands/help.go, cmd/plumego/internal/checker/checker.go, cmd/plumego/internal/checker/checker_test.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: none

Goal:
Make `plumego check` deterministic by default and move network/update checks behind an explicit flag.

Scope:
- Add `check --updates`.
- Keep `go mod verify` in the default dependency check.
- Run `go list -u -m all` only when `--updates` is set.
- Update help metadata and regression tests.

Non-goals:
- Do not change security/config/structure check semantics.
- Do not add module proxy configuration.
- Do not change check output schema beyond omitting `outdated` when updates are not requested.

Files:
- `cmd/plumego/commands/check.go`
- `cmd/plumego/commands/help.go`
- `cmd/plumego/internal/checker/checker.go`
- `cmd/plumego/internal/checker/checker_test.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./internal/checker ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required unless README check examples change.

Done Definition:
- Default check does not invoke `go list -u -m all`.
- `--updates` preserves update warning behavior.
- Help lists `--updates`.

Outcome:
