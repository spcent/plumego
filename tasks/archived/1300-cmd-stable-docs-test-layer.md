# Card 1300

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P2
State: done
Primary Module: cmd/plumego stable docs/tests
Owned Files: cmd/plumego/README.md, cmd/plumego/MODULE.md, cmd/plumego/commands/root_help_test.go, cmd/plumego/commands/project_smoke_test.go, cmd/plumego/internal/routeanalyzer/analyzer.go
Depends On: 0767

Goal:
Clarify stable boundaries for route analysis, slow smoke tests, and CLI binary artifacts.

Scope:
- Document route analyzer as static best-effort direct literal route extraction.
- Document `go test -short ./commands` as fast contract gate and full tests as slow smoke.
- Document CLI build artifact location outside `cmd/plumego/plumego`.
- Add lightweight tests for route analyzer boundary wording where practical.

Non-goals:
- Do not implement group/middleware extraction.
- Do not remove the ignored local binary file.
- Do not change command behavior.

Files:
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`
- `cmd/plumego/commands/root_help_test.go`
- `cmd/plumego/commands/project_smoke_test.go`
- `cmd/plumego/internal/routeanalyzer/analyzer.go`

Tests:
- `go test -short ./commands ./internal/routeanalyzer`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`

Done Definition:
- Stable docs explain current analyzer/test/artifact boundaries.
- Fast and slow command test layers are discoverable.

Outcome:
- Clarified the route analyzer stable boundary as static best-effort direct literal route extraction.
- Documented fast command-contract testing versus the slow generated-project smoke layer.
- Documented the supported CLI build artifact path under repository-level `bin/plumego` and warned against stale `cmd/plumego/plumego` binaries.
- Added documentation guard tests for route analyzer boundaries, test layers, and build artifact guidance.

Validation:
- `go test -short ./commands ./internal/routeanalyzer`
- `go test ./...`
- `go vet ./...`
