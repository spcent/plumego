# Card 0999

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/inspect.go, cmd/plumego/commands/routes.go, cmd/plumego/internal/routeanalyzer/analyzer.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0740

Goal:
Harden runtime inspect limits and make route analyzer feature boundaries honest.

Scope:
- Detect inspect response truncation and return a clear structured error.
- Clarify `--auth` as an Authorization header value or implement documented token behavior.
- Remove or hard-fail unsupported route analyzer flags consistently.
- Validate `--sort` values instead of silently defaulting.
- Add focused tests for truncation, auth header behavior, unsupported group, and invalid sort.

Non-goals:
- Do not add dynamic runtime route discovery.
- Do not implement full Go data-flow analysis for route extraction.
- Do not expose new debug endpoints.

Files:
- `cmd/plumego/commands/inspect.go`
- `cmd/plumego/commands/routes.go`
- `cmd/plumego/internal/routeanalyzer/analyzer.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands ./internal/routeanalyzer`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Oversized inspect responses fail with an explicit limit error.
- Auth flag behavior is tested and documented.
- Routes command rejects unsupported/invalid analyzer options predictably.

Outcome:
- Added explicit inspect response limit detection instead of silent truncation.
- Tested `--auth` as a full Authorization header value.
- Added route analyzer sort validation and rejected unsupported `group` sorting.
- Documented inspect limit/auth behavior and static route analyzer boundaries.

Validation:
- `go test ./commands ./internal/routeanalyzer`
- `go test ./...`
- `go vet ./...`
