# Card 1272

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
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
- Documented the two machine output shapes in `cmd/plumego/README.md`: command-result envelopes for single command results and direct event objects for streaming updates.
- Added JSON/YAML/text command-result contract coverage.
- Added JSON/YAML/text event contract coverage and assertions that event output is not wrapped in command-result envelopes.
- Kept raw string `Print` output enveloped in JSON/YAML.

Validation:
- `go test ./internal/output`
- `go test ./...`
- `go vet ./...`
