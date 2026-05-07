# Card 0716

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/config.go, cmd/plumego/internal/configmgr/configmgr.go, cmd/plumego/internal/configmgr/configmgr_test.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0715

Goal:
Make configuration commands fail closed on secret disclosure.

Scope:
- Redact sensitive configuration values by default.
- Require an explicit opt-in flag for raw secret display.
- Ensure environment output keeps sensitive values redacted.
- Add tests for default redaction and explicit opt-in behavior.

Non-goals:
- Do not redesign configuration loading.
- Do not change generated application config semantics.
- Do not add secret scanning dependencies.

Files:
- `cmd/plumego/commands/config.go`
- `cmd/plumego/internal/configmgr/configmgr.go`
- `cmd/plumego/internal/configmgr/configmgr_test.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands ./internal/configmgr`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md` if user-facing flags change.

Done Definition:
- `plumego config show` does not leak secrets by default.
- Tests prove redaction behavior.

Outcome:
- Made `plumego config show` redact sensitive values by default.
- Added explicit `--show-secrets` opt-in for trusted local raw-value output.
- Reworked config redaction to apply recursively across nested maps.
- Documented the default redaction behavior in the CLI README.
- Added command-level and configmgr redaction regression tests.
- Validation Run:
  - `go test ./commands ./internal/configmgr`
  - `go build .`
