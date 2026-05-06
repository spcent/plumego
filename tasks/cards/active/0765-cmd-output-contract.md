# Card 0765

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: active
Primary Module: cmd/plumego output
Owned Files: cmd/plumego/internal/output/formatter.go, cmd/plumego/internal/output/event.go, cmd/plumego/internal/output/formatter_test.go, cmd/plumego/README.md
Depends On: 0764

Goal:
Make the CLI output contract explicit and covered by tests.

Scope:
- Document result-envelope output versus streaming event output.
- Add contract tests for JSON/YAML/text command results and JSON/YAML/text events.
- Ensure raw string printing remains enveloped in JSON/YAML.

Non-goals:
- Do not collapse streaming events into command result envelopes.
- Do not change existing event field names.
- Do not redesign text output.

Files:
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/internal/output/event.go`
- `cmd/plumego/internal/output/formatter_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/output`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Machine consumers have documented parsing expectations.
- Output contract tests cover both result and event forms.

Outcome:
