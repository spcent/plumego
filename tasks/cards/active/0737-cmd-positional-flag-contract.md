# Card 0737

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/new.go, cmd/plumego/commands/build.go, cmd/plumego/commands/generate.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0736

Goal:
Normalize high-risk positional flag parsing so documented CLI examples work and extra arguments fail closed.

Scope:
- Make `plumego new myapp --template api` and similar documented examples parse correctly.
- Reject unexpected extra positional arguments in `new`, `build`, and `generate`.
- Add regression tests for flags after positional arguments and extra-argument failures.
- Sync README only if command usage text changes.

Non-goals:
- Do not change global flag parsing order.
- Do not change generated project contents.
- Do not touch dev server, migration, or route analyzer behavior.

Files:
- `cmd/plumego/commands/new.go`
- `cmd/plumego/commands/build.go`
- `cmd/plumego/commands/generate.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Documented `new` examples parse correctly with flags after the project name.
- `new`, `build`, and `generate` reject unexpected positional tails.
- Focused CLI tests pass.

