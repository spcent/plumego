# Card 2160: cmd/plumego Output Result Envelope Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/internal/output/formatter_test.go`
Depends On: none

Goal:
Converge CLI success and error output on a typed command-result envelope.

Problem:
`Formatter.Success` and `Formatter.Error` still build their JSON/YAML/text
payloads with mutable `map[string]any` values. This hides the intended CLI
field contract and duplicates envelope construction.

Scope:
- Add a local typed result envelope for CLI command output.
- Keep existing field names: `status`, `message`, `exit_code`, and `data`.
- Add focused tests for JSON success and error output.

Non-goals:
- Do not change HTTP response contracts.
- Do not change event output, text rendering, or command behavior.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/internal/output/formatter_test.go`

Tests:
- `go test -race -timeout 60s ./internal/output/...` from `cmd/plumego`
- `go test -timeout 20s ./internal/output/...` from `cmd/plumego`
- `go vet ./internal/output/...` from `cmd/plumego`

Docs Sync:
No docs change required; output fields remain unchanged.

Done Definition:
- Success/error envelopes no longer use one-off maps.
- Focused output tests cover typed JSON success and error output.
- The listed validation commands pass.

Outcome:
