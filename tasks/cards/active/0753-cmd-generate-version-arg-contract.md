# Card 0753

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego argument contract
Owned Files: cmd/plumego/commands/generate.go, cmd/plumego/commands/version.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0752

Goal:
Align remaining command argument behavior with the stable CLI contract.

Scope:
- Add `--dir` support to `generate`.
- Resolve generate output relative to the selected project directory.
- Reject unexpected positional arguments for `version`.
- Add CLI regression tests for `generate --dir` and `version extra`.

Non-goals:
- Do not change generated code templates.
- Do not add new generator types.
- Do not change version payload fields.

Files:
- `cmd/plumego/commands/generate.go`
- `cmd/plumego/commands/version.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required beyond help metadata from 0752.

Done Definition:
- `plumego generate --dir <path>` writes into the selected project.
- `plumego version extra` fails with a structured error.

Outcome:
