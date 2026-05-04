# Card 0724

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/scaffold
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go, cmd/plumego/commands/new.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0723

Goal:
Make generated project dependency files truthful and reproducible.

Scope:
- Stop running `go mod init` after writing a generated `go.mod`.
- Make generated `go.mod` content explicit about local development versus released module versions.
- Add tests that generated canonical and API projects have the expected module file shape.
- Keep generated projects buildable in repository-local test scenarios.

Non-goals:
- Do not introduce a package manager or template engine.
- Do not redesign every scaffold profile.
- Do not change stable core APIs.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/commands/new.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/scaffold ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- `plumego new` does not silently execute redundant module initialization.
- Generated module files have deterministic, documented content.
- Scaffold tests cover the generated `go.mod` contract.

Outcome:
- Stopped running `go mod init`/`go mod tidy` after writing scaffolded
  `go.mod` files.
- Added explicit scaffold project options for generated Plumego requirement and
  optional local `replace` directives.
- Wired `plumego new` to detect a local Plumego checkout and include a local
  replace for source-checkout workflows.
- Added scaffold and CLI regression tests for generated `go.mod` content.
- Validation Run:
  - `go test ./internal/scaffold ./commands`
  - `go build .`
