# Card 0726

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/test.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0725

Goal:
Make `plumego test` reports trustworthy for CI and code agents.

Scope:
- Avoid overwriting a user's `coverage.out` unless explicitly requested.
- Compute coverage through `go tool cover -func` or equivalent accurate parsing.
- Preserve package-level failures and stderr in error output.
- Add CLI tests for failing tests and coverage reporting.

Non-goals:
- Do not replace `go test`.
- Do not add external test-report dependencies.
- Do not implement junit output.

Files:
- `cmd/plumego/commands/test.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- `plumego test --cover` does not clobber default user files.
- Coverage percentage matches `go tool cover -func`.
- Failing package output remains visible in structured error data.

Outcome:
- Changed `plumego test --cover` to use a temporary coverage profile unless
  `--coverprofile` is explicitly provided.
- Switched coverage calculation to `go tool cover -func` with the tested project
  as the working directory.
- Preserved structured test and package failures, plus stderr when available, in
  failure output data.
- Added CLI regression tests for coverage and failing-test reports.
- Validation Run:
  - `go test ./commands`
  - `go build .`
