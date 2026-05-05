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
