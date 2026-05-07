# Card 0767

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego devserver concurrency
Owned Files: cmd/plumego/internal/devserver/deps.go, cmd/plumego/internal/devserver/deps_test.go, cmd/plumego/internal/devserver/runner.go, cmd/plumego/internal/devserver/runner_test.go
Depends On: 0766

Goal:
Avoid dependency graph rebuild storms and harden runner log scanning.

Scope:
- Coalesce concurrent dependency graph refreshes so only one `go list` graph build runs per cache miss.
- Keep stale cached graph serving semantics unchanged.
- Increase or remove the `bufio.Scanner` token limit for app stdout/stderr.
- Add concurrency/log-line regression tests.

Non-goals:
- Do not add `x/sync/singleflight`.
- Do not redesign dependency graph payloads.
- Do not change runner stop ownership.

Files:
- `cmd/plumego/internal/devserver/deps.go`
- `cmd/plumego/internal/devserver/deps_test.go`
- `cmd/plumego/internal/devserver/runner.go`
- `cmd/plumego/internal/devserver/runner_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Concurrent cache misses share one in-flight dependency graph build.
- Long single-line app logs do not break output streaming.

Outcome:
- Added in-flight dependency graph refresh coalescing to `depsCache` so concurrent cache misses share one build.
- Preserved existing cached graph return semantics for fresh non-refresh reads.
- Increased runner scanner capacity to handle long single-line app logs.
- Added regression tests for concurrent dependency graph cache misses and long app log lines.

Validation:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`
