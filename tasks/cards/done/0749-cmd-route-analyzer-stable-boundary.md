# Card 0749

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: active
Primary Module: cmd/plumego route analyzer
Owned Files: cmd/plumego/commands/inspect.go, cmd/plumego/commands/routes.go, cmd/plumego/internal/routeanalyzer/analyzer.go, cmd/plumego/internal/routeanalyzer/analyzer_test.go
Depends On: 0748

Goal:
Make route analyzer behavior match its stable CLI contract.

Scope:
- Remove or clearly suppress unimplemented middleware/group fields from command output.
- Keep analyzer best-effort behavior explicit in code and tests.
- Add regression tests for unsupported group/middleware requests.

Non-goals:
- Do not build a full Go interpreter or router execution model.
- Do not add dependencies.
- Do not change core/router APIs.

Files:
- `cmd/plumego/commands/inspect.go`
- `cmd/plumego/commands/routes.go`
- `cmd/plumego/internal/routeanalyzer/analyzer.go`
- `cmd/plumego/internal/routeanalyzer/analyzer_test.go`

Tests:
- `go test ./commands ./internal/routeanalyzer`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless README command examples change.

Done Definition:
- Route analyzer output does not imply unimplemented group/middleware extraction.
- Unsupported flags fail closed or emit documented empty behavior.
- Tests lock the stable best-effort boundary.

Outcome:
- Removed unsupported group and middleware fields from the static route analyzer
  result model.
- Kept `--group` as an explicit fail-closed unsupported option and changed
  `--middleware` to fail closed instead of returning an empty implied summary.
- Updated routes help text and sort flag wording to advertise only supported
  analyzer behavior.
- Added regression tests for unsupported middleware/group/sort options and for
  JSON output omitting unsupported route metadata fields.

Validation:
- `go test ./commands ./internal/routeanalyzer` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
