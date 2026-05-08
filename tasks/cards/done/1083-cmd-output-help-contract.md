# Card 1083

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: active
Primary Module: cmd/plumego output and help
Owned Files: cmd/plumego/commands/root.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/internal/output/formatter.go, cmd/plumego/internal/output/formatter_test.go
Depends On: 0747

Goal:
Remove raw machine-output string hazards and reduce help metadata drift.

Scope:
- Make JSON/YAML string output use a structured value or fail closed.
- Add contract tests for help, error, warning, and success output envelopes.
- Consolidate help metadata enough that command-specific help cannot diverge silently.

Non-goals:
- Do not redesign the CLI command interface.
- Do not add generated documentation.
- Do not change text output beyond contract fixes.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/internal/output/formatter_test.go`

Tests:
- `go test ./commands ./internal/output`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless help behavior changes visibly.

Done Definition:
- JSON/YAML output never emits bare raw strings.
- Help and output envelopes have regression coverage.
- Help metadata drift is reduced or explicitly guarded by tests.

Outcome:
- Wrapped JSON/YAML `Formatter.Print(string)` output in the command-result
  envelope instead of emitting raw strings.
- Preserved text-mode direct string printing for command help readability.
- Added formatter regression tests for success, warning, error, JSON string,
  YAML string, and text command-result output.
- Added CLI regression coverage for YAML command help envelopes.

Validation:
- `go test ./commands ./internal/output` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
