# Card 0751

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P2
State: active
Primary Module: cmd/plumego test and artifact hygiene
Owned Files: cmd/plumego/README.md, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/commands/dev_test.go, .gitignore
Depends On: 0750

Goal:
Make cmd validation and local build artifacts easier to trust before stable.

Scope:
- Separate slow CLI smoke coverage from fast command contract tests where practical.
- Avoid global cwd changes where a test can use command context directly.
- Document or redirect local CLI binary build artifacts.

Non-goals:
- Do not delete existing coverage.
- Do not introduce a new test framework.
- Do not change release packaging.

Files:
- `cmd/plumego/README.md`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/commands/dev_test.go`
- `.gitignore`

Tests:
- `go test ./commands`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Fast command contract tests remain usable in normal development.
- Slow smoke tests are identifiable.
- Local binary artifact handling is explicit.

Outcome:
- Marked the generated-project CLI stable workflow as a slow smoke path that is
  skipped by `testing.Short()`.
- Documented fast command-contract testing with `go test -short ./commands` and
  full CLI confidence with `go test ./...` from `cmd/plumego`.
- Documented repository-level `bin/` as the normal local CLI binary target and
  kept source-directory binaries ignored only as cleanup guards.
- Expanded `.gitignore` coverage for accidental `cmd/plumego/plumego-*` local
  binaries.

Validation:
- `go test -short ./commands` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
