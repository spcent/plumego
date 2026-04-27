# Card 0461: cmd/plumego CLI E2E JSON DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/commands/cli_e2e_test.go`
Depends On: none

Goal:
Make fixed-shape CLI JSON output tests decode into typed DTOs instead of
generic maps.

Problem:
Several CLI E2E tests validate stable JSON envelopes and inspect output fields
through `map[string]any`. This makes the expected command output shape harder to
see and differs from the typed CLI output tests already present in the same
file.

Scope:
- Add small local DTOs for the common CLI JSON envelope and inspect responses.
- Replace fixed-shape map decoding in version and inspect E2E assertions.
- Leave genuinely dynamic config maps unchanged.

Non-goals:
- Do not change command behavior or output format.
- Do not add dependencies.

Files:
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- from `cmd/plumego`: `go test -race -timeout 60s ./commands/...`
- from `cmd/plumego`: `go test -timeout 20s ./commands/...`
- from `cmd/plumego`: `go vet ./commands/...`

Docs Sync:
No docs change required; this is test-only output DTO convergence.

Done Definition:
- Fixed-shape CLI E2E JSON output assertions no longer decode through
  `map[string]any`.
- The listed validation commands pass.

Outcome:
