# Card 0750

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: active
Primary Module: cmd/plumego devserver performance
Owned Files: cmd/plumego/internal/devserver/builder.go, cmd/plumego/internal/devserver/builder_test.go, cmd/plumego/internal/devserver/deps.go, cmd/plumego/internal/devserver/deps_test.go
Depends On: 0749

Goal:
Reduce devserver memory and concurrency hazards.

Scope:
- Bound captured build stdout/stderr before publishing build events.
- Keep dependency graph cache locks out of `go list` execution.
- Return validation errors for invalid dependency graph query parameters.

Non-goals:
- Do not change dashboard route URLs.
- Do not replace `go list`.
- Do not add async cache refresh.

Files:
- `cmd/plumego/internal/devserver/builder.go`
- `cmd/plumego/internal/devserver/builder_test.go`
- `cmd/plumego/internal/devserver/deps.go`
- `cmd/plumego/internal/devserver/deps_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless dashboard API response semantics change.

Done Definition:
- Build output cannot grow unbounded in memory or event payloads.
- Dependency cache callers are not blocked by a long lock around `go list`.
- Invalid `max_nodes` values return a clear error.

Outcome:
- Added bounded build output capture that continues consuming command output
  while retaining only a fixed prefix plus a truncation marker.
- Updated dependency graph cache reads to return fresh cache hits under lock but
  execute `go list` graph construction outside the mutex.
- Rejected invalid `max_nodes` query values with a validation error instead of
  silently treating them as unlimited.
- Added devserver tests for bounded output capture and invalid dependency graph
  query parameters.

Validation:
- `go test ./internal/devserver` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
