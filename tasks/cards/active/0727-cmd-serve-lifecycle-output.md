# Card 0727

Milestone: cmd stable hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/serve.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0726

Goal:
Bring `plumego serve` into the same lifecycle and output contract as the rest of the CLI.

Scope:
- Use `http.Server` with signal-aware graceful shutdown.
- Validate directory and address flags consistently.
- Emit structured start/stop/error output.
- Add tests for invalid directory and help/output behavior.

Non-goals:
- Do not add TLS or production static hosting features.
- Do not make `serve` part of the app scaffold runtime.
- Do not introduce a router dependency.

Files:
- `cmd/plumego/commands/serve.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- `serve` no longer owns a raw blocking `ListenAndServe` path without shutdown.
- CLI output remains machine-readable outside text mode.
- Focused command tests cover error and help behavior.
