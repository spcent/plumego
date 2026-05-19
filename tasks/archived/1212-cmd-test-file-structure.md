# Card 1212

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P2
State: done
Primary Module: cmd/plumego tests
Owned Files: cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/commands/root_help_test.go, cmd/plumego/commands/project_smoke_test.go, cmd/plumego/commands/config_cli_test.go
Depends On: 0758

Goal:
Improve command test maintainability without reducing coverage.

Scope:
- Move root/help contract tests into a dedicated file.
- Move generated project smoke tests into a dedicated file.
- Move config CLI contract tests into a dedicated file where low-risk.
- Keep shared helpers stable and avoid changing test behavior.

Non-goals:
- Do not rewrite all command tests.
- Do not remove slow smoke coverage.
- Do not add a test framework.

Files:
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/commands/root_help_test.go`
- `cmd/plumego/commands/project_smoke_test.go`
- `cmd/plumego/commands/config_cli_test.go`

Tests:
- `go test -short ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Large command e2e file is reduced by moving focused groups out.
- Test behavior remains unchanged.
- Short and full command tests pass.

Outcome:
- Split root/help contract tests into `root_help_test.go`.
- Split config CLI contract tests into `config_cli_test.go`.
- Split generated project and `new` command smoke tests into
  `project_smoke_test.go`.
- Kept shared helpers in `cli_e2e_test.go`; test behaviour is unchanged.

Validation:
- `go test -short ./commands`
- `go test ./...`
- `go vet ./...`
