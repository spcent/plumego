# Card 0755

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego devserver deps
Owned Files: cmd/plumego/internal/devserver/deps.go, cmd/plumego/internal/devserver/dashboard_info_test.go, cmd/plumego/internal/devserver/deps_test.go
Depends On: 0754

Goal:
Remove dependency graph `go list` stderr pipe deadlock risk.

Scope:
- Drain stderr concurrently or use a bounded combined-output approach while decoding stdout.
- Bound dependency graph diagnostic output.
- Add focused tests for stderr capture and invalid output handling.

Non-goals:
- Do not replace `go list`.
- Do not change dependency graph response schema except bounded error text.
- Do not add async refresh.

Files:
- `cmd/plumego/internal/devserver/deps.go`
- `cmd/plumego/internal/devserver/deps_test.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- `go list` stderr cannot block dependency graph collection.
- Error text is bounded and safe.
- Tests cover stderr handling.

Outcome:
